package itest

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/channeldb"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/lnwallet"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name string
	test func(net *lntest.NetworkHarness, t *harnessTest)
}

// harnessTest wraps a regular testing.T providing enhanced error detection
// and propagation. All error will be augmented with a full stack-trace in
// order to aid in debugging. Additionally, any panics caused by active
// test cases will also be handled and represented as fatals.
type harnessTest struct {
	t *testing.T

	// testCase is populated during test execution and represents the
	// current test case.
	testCase *testCase

	// lndHarness is a reference to the current network harness. Will be
	// nil if not yet set up.
	lndHarness *lntest.NetworkHarness
}

// newHarnessTest creates a new instance of a harnessTest from a regular
// testing.T instance.
func newHarnessTest(t *testing.T, net *lntest.NetworkHarness) *harnessTest {
	return &harnessTest{t, nil, net}
}

// Skipf calls the underlying testing.T's Skip method, causing the current test
// to be skipped.
func (h *harnessTest) Skipf(format string, args ...interface{}) {
	h.t.Skipf(format, args...)
}

// Fatalf causes the current active test case to fail with a fatal error. All
// integration tests should mark test failures solely with this method due to
// the error stack traces it produces.
func (h *harnessTest) Fatalf(format string, a ...interface{}) {
	if h.lndHarness != nil {
		h.lndHarness.SaveProfilesPages(h.t)
	}

	stacktrace := errors.Wrap(fmt.Sprintf(format, a...), 1).ErrorStack()

	if h.testCase != nil {
		h.t.Fatalf("Failed: (%v): exited with error: \n"+
			"%v", h.testCase.name, stacktrace)
	} else {
		h.t.Fatalf("Error outside of test: %v", stacktrace)
	}
}

// RunTestCase executes a harness test case. Any errors or panics will be
// represented as fatal.
func (h *harnessTest) RunTestCase(testCase *testCase) {
	h.testCase = testCase
	defer func() {
		h.testCase = nil
	}()

	defer func() {
		if err := recover(); err != nil {
			description := errors.Wrap(err, 2).ErrorStack()
			h.t.Fatalf("Failed: (%v) panicked with: \n%v",
				h.testCase.name, description)
		}
	}()

	testCase.test(h.lndHarness, h)
}

func (h *harnessTest) Logf(format string, args ...interface{}) {
	h.t.Logf(format, args...)
}

func (h *harnessTest) Log(args ...interface{}) {
	h.t.Log(args...)
}

func assertTxInBlock(t *harnessTest, block *wire.MsgBlock, txid *chainhash.Hash) {
	for _, tx := range block.Transactions {
		sha := tx.TxHash()
		if bytes.Equal(txid[:], sha[:]) {
			return
		}
	}

	t.Fatalf("tx was not included in block")
}

// mineBlocks mine 'num' of blocks and check that blocks are present in
// node blockchain. numTxs should be set to the number of transactions
// (excluding the coinbase) we expect to be included in the first mined block.
func mineBlocks(t *harnessTest, net *lntest.NetworkHarness,
	num uint32, numTxs int) []*wire.MsgBlock {

	// If we expect transactions to be included in the blocks we'll mine,
	// we wait here until they are seen in the miner's mempool.
	var txids []*chainhash.Hash
	var err error
	if numTxs > 0 {
		txids, err = waitForNTxsInMempool(
			net.Miner.Client, numTxs, minerMempoolTimeout,
		)
		if err != nil {
			t.Fatalf("unable to find txns in mempool: %v", err)
		}
	}

	blocks := make([]*wire.MsgBlock, num)

	blockHashes, err := net.Miner.Client.Generate(num)
	if err != nil {
		t.Fatalf("unable to generate blocks: %v", err)
	}

	for i, blockHash := range blockHashes {
		block, err := net.Miner.Client.GetBlock(blockHash)
		if err != nil {
			t.Fatalf("unable to get block: %v", err)
		}

		blocks[i] = block
	}

	// Finally, assert that all the transactions were included in the first
	// block.
	for _, txid := range txids {
		assertTxInBlock(t, blocks[0], txid)
	}

	return blocks
}

// openChannelStream blocks until an OpenChannel request for a channel funding
// by alice succeeds. If it does, a stream client is returned to receive events
// about the opening channel.
func openChannelStream(t *harnessTest, net *lntest.NetworkHarness,
	alice, bob *lntest.HarnessNode,
	p lntest.OpenChannelParams) lnrpc.Lightning_OpenChannelClient {

	t.t.Helper()

	// Wait until we are able to fund a channel successfully. This wait
	// prevents us from erroring out when trying to create a channel while
	// the node is starting up.
	var chanOpenUpdate lnrpc.Lightning_OpenChannelClient
	err := wait.NoError(func() error {
		var err error
		chanOpenUpdate, err = net.OpenChannel(alice, bob, p)
		return err
	}, defaultTimeout)
	if err != nil {
		t.Fatalf("unable to open channel: %v", err)
	}

	return chanOpenUpdate
}

// openChannelAndAssert attempts to open a channel with the specified
// parameters extended from Alice to Bob. Additionally, two items are asserted
// after the channel is considered open: the funding transaction should be
// found within a block, and that Alice can report the status of the new
// channel.
func openChannelAndAssert(t *harnessTest, net *lntest.NetworkHarness,
	alice, bob *lntest.HarnessNode,
	p lntest.OpenChannelParams) *lnrpc.ChannelPoint {

	t.t.Helper()

	chanOpenUpdate := openChannelStream(t, net, alice, bob, p)

	// Mine 6 blocks, then wait for Alice's node to notify us that the
	// channel has been opened. The funding transaction should be found
	// within the first newly mined block. We mine 6 blocks so that in the
	// case that the channel is public, it is announced to the network.
	block := mineBlocks(t, net, 6, 1)[0]

	fundingChanPoint, err := net.WaitForChannelOpen(chanOpenUpdate)
	if err != nil {
		t.Fatalf("error while waiting for channel open: %v", err)
	}
	fundingTxID, err := lnrpc.GetChanPointFundingTxid(fundingChanPoint)
	if err != nil {
		t.Fatalf("unable to get txid: %v", err)
	}
	assertTxInBlock(t, block, fundingTxID)

	// The channel should be listed in the peer information returned by
	// both peers.
	chanPoint := wire.OutPoint{
		Hash:  *fundingTxID,
		Index: fundingChanPoint.OutputIndex,
	}
	if err := net.AssertChannelExists(alice, &chanPoint); err != nil {
		t.Fatalf("unable to assert channel existence: %v", err)
	}
	if err := net.AssertChannelExists(bob, &chanPoint); err != nil {
		t.Fatalf("unable to assert channel existence: %v", err)
	}

	return fundingChanPoint
}

// closeChannelAndAssert attempts to close a channel identified by the passed
// channel point owned by the passed Lightning node. A fully blocking channel
// closure is attempted, therefore the passed context should be a child derived
// via timeout from a base parent. Additionally, once the channel has been
// detected as closed, an assertion checks that the transaction is found within
// a block. Finally, this assertion verifies that the node always sends out a
// disable update when closing the channel if the channel was previously enabled.
//
// NOTE: This method assumes that the provided funding point is confirmed
// on-chain AND that the edge exists in the node's channel graph. If the funding
// transactions was reorged out at some point, use closeReorgedChannelAndAssert.
func closeChannelAndAssert(t *harnessTest, net *lntest.NetworkHarness,
	node *lntest.HarnessNode, fundingChanPoint *lnrpc.ChannelPoint,
	force bool) *chainhash.Hash {

	return closeChannelAndAssertType(t, net, node, fundingChanPoint, false, force)
}

func closeChannelAndAssertType(t *harnessTest, net *lntest.NetworkHarness,
	node *lntest.HarnessNode, fundingChanPoint *lnrpc.ChannelPoint,
	anchors, force bool) *chainhash.Hash {

	ctxb := context.Background()
	ctxt, cancel := context.WithTimeout(ctxb, channelCloseTimeout)
	defer cancel()

	// Fetch the current channel policy. If the channel is currently
	// enabled, we will register for graph notifications before closing to
	// assert that the node sends out a disabling update as a result of the
	// channel being closed.
	curPolicy := getChannelPolicies(t, node, node.PubKeyStr, fundingChanPoint)[0]
	expectDisable := !curPolicy.Disabled

	closeUpdates, _, err := net.CloseChannel(node, fundingChanPoint, force)
	require.NoError(t.t, err, "unable to close channel")

	// If the channel policy was enabled prior to the closure, wait until we
	// received the disabled update.
	if expectDisable {
		curPolicy.Disabled = true
		assertChannelPolicyUpdate(
			t.t, node, node.PubKeyStr,
			curPolicy, fundingChanPoint, false,
		)
	}

	return assertChannelClosed(
		ctxt, t, net, node, fundingChanPoint, anchors, closeUpdates,
	)
}

// assertChannelPolicyUpdate checks that the required policy update has
// happened on the given node.
func assertChannelPolicyUpdate(t *testing.T, node *lntest.HarnessNode,
	advertisingNode string, policy *lnrpc.RoutingPolicy,
	chanPoint *lnrpc.ChannelPoint, includeUnannounced bool) {

	ctxb := context.Background()
	ctxt, cancel := context.WithTimeout(ctxb, lntest.DefaultTimeout)
	defer cancel()

	require.NoError(
		t, node.WaitForChannelPolicyUpdate(
			ctxt, advertisingNode, policy,
			chanPoint, includeUnannounced,
		), "error while waiting for channel update",
	)
}

// assertChannelClosed asserts that the channel is properly cleaned up after
// initiating a cooperative or local close.
func assertChannelClosed(ctx context.Context, t *harnessTest,
	net *lntest.NetworkHarness, node *lntest.HarnessNode,
	fundingChanPoint *lnrpc.ChannelPoint, anchors bool,
	closeUpdates lnrpc.Lightning_CloseChannelClient) *chainhash.Hash {

	txid, err := lnrpc.GetChanPointFundingTxid(fundingChanPoint)
	if err != nil {
		t.Fatalf("unable to get txid: %v", err)
	}
	chanPointStr := fmt.Sprintf("%v:%v", txid, fundingChanPoint.OutputIndex)

	// If the channel appears in list channels, ensure that its state
	// contains ChanStatusCoopBroadcasted.
	ctxt, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	listChansRequest := &lnrpc.ListChannelsRequest{}
	listChansResp, err := node.ListChannels(ctxt, listChansRequest)
	if err != nil {
		t.Fatalf("unable to query for list channels: %v", err)
	}
	for _, channel := range listChansResp.Channels {
		// Skip other channels.
		if channel.ChannelPoint != chanPointStr {
			continue
		}

		// Assert that the channel is in coop broadcasted.
		if !strings.Contains(channel.ChanStatusFlags,
			channeldb.ChanStatusCoopBroadcasted.String()) {
			t.Fatalf("channel not coop broadcasted, "+
				"got: %v", channel.ChanStatusFlags)
		}
	}

	// At this point, the channel should now be marked as being in the
	// state of "waiting close".
	ctxt, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	pendingChansRequest := &lnrpc.PendingChannelsRequest{}
	pendingChanResp, err := node.PendingChannels(ctxt, pendingChansRequest)
	if err != nil {
		t.Fatalf("unable to query for pending channels: %v", err)
	}
	var found bool
	for _, pendingClose := range pendingChanResp.WaitingCloseChannels {
		if pendingClose.Channel.ChannelPoint == chanPointStr {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("channel not marked as waiting close")
	}

	// We'll now, generate a single block, wait for the final close status
	// update, then ensure that the closing transaction was included in the
	// block. If there are anchors, we also expect an anchor sweep.
	expectedTxes := 1
	if anchors {
		expectedTxes = 2
	}

	block := mineBlocks(t, net, 1, expectedTxes)[0]

	closingTxid, err := net.WaitForChannelClose(closeUpdates)
	if err != nil {
		t.Fatalf("error while waiting for channel close: %v", err)
	}

	assertTxInBlock(t, block, closingTxid)

	// Finally, the transaction should no longer be in the waiting close
	// state as we've just mined a block that should include the closing
	// transaction.
	err = wait.Predicate(func() bool {
		pendingChansRequest := &lnrpc.PendingChannelsRequest{}
		pendingChanResp, err := node.PendingChannels(
			ctx, pendingChansRequest,
		)
		if err != nil {
			return false
		}

		for _, pendingClose := range pendingChanResp.WaitingCloseChannels {
			if pendingClose.Channel.ChannelPoint == chanPointStr {
				return false
			}
		}

		return true
	}, time.Second*15)
	if err != nil {
		t.Fatalf("closing transaction not marked as fully closed")
	}

	return closingTxid
}

// shutdownAndAssert shuts down the given node and asserts that no errors
// occur.
func shutdownAndAssert(net *lntest.NetworkHarness, t *harnessTest,
	node *lntest.HarnessNode) {
	if err := net.ShutdownNode(node); err != nil {
		t.Fatalf("unable to shutdown %v: %v", node.Name(), err)
	}
}

// calcStaticFee calculates appropriate fees for commitment transactions.  This
// function provides a simple way to allow test balance assertions to take fee
// calculations into account.
func calcStaticFee(c lnrpc.CommitmentType, numHTLCs int) btcutil.Amount {
	const htlcWeight = input.HTLCWeight
	var (
		feePerKw     = chainfee.SatPerKVByte(50000).FeePerKWeight()
		commitWeight = input.CommitWeight
		anchors      = btcutil.Amount(0)
	)

	// The anchor commitment type is slightly heavier, and we must also add
	// the value of the two anchors to the resulting fee the initiator
	// pays. In addition the fee rate is capped at 10 sat/vbyte for anchor
	// channels.
	if commitTypeHasAnchors(c) {
		feePerKw = chainfee.SatPerKVByte(
			lnwallet.DefaultAnchorsCommitMaxFeeRateSatPerVByte * 1000,
		).FeePerKWeight()
		commitWeight = input.AnchorCommitWeight
		anchors = 2 * anchorSize
	}

	return feePerKw.FeeForWeight(int64(commitWeight+htlcWeight*numHTLCs)) +
		anchors
}

// commitTypeHasAnchors returns whether commitType uses anchor outputs.
func commitTypeHasAnchors(commitType lnrpc.CommitmentType) bool {
	switch commitType {
	case lnrpc.CommitmentType_ANCHORS,
		lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE:
		return true
	default:
		return false
	}
}

// channelCommitType retrieves the active channel commitment type for the given
// chan point.
func channelCommitType(node *lntest.HarnessNode,
	chanPoint *lnrpc.ChannelPoint) (lnrpc.CommitmentType, error) {

	ctxb := context.Background()
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()

	req := &lnrpc.ListChannelsRequest{}
	channels, err := node.ListChannels(ctxt, req)
	if err != nil {
		return 0, fmt.Errorf("listchannels failed: %v", err)
	}

	for _, c := range channels.Channels {
		if c.ChannelPoint == txStr(chanPoint) {
			return c.CommitmentType, nil
		}
	}

	return 0, fmt.Errorf("channel point %v not found", chanPoint)
}

// txStr returns the string representation of the channel's funding transaction.
func txStr(chanPoint *lnrpc.ChannelPoint) string {
	fundingTxID, err := lnrpc.GetChanPointFundingTxid(chanPoint)
	if err != nil {
		return ""
	}
	cp := wire.OutPoint{
		Hash:  *fundingTxID,
		Index: chanPoint.OutputIndex,
	}
	return cp.String()
}

// getChannelPolicies queries the channel graph and retrieves the current edge
// policies for the provided channel points.
func getChannelPolicies(t *harnessTest, node *lntest.HarnessNode,
	advertisingNode string,
	chanPoints ...*lnrpc.ChannelPoint) []*lnrpc.RoutingPolicy {

	ctxb := context.Background()

	descReq := &lnrpc.ChannelGraphRequest{
		IncludeUnannounced: true,
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	chanGraph, err := node.DescribeGraph(ctxt, descReq)
	require.NoError(t.t, err, "unable to query for alice's graph")

	var policies []*lnrpc.RoutingPolicy
	err = wait.NoError(func() error {
	out:
		for _, chanPoint := range chanPoints {
			for _, e := range chanGraph.Edges {
				if e.ChanPoint != txStr(chanPoint) {
					continue
				}

				if e.Node1Pub == advertisingNode {
					policies = append(policies, e.Node1Policy)
				} else {
					policies = append(policies, e.Node2Policy)
				}

				continue out
			}

			// If we've iterated over all the known edges and we weren't
			// able to find this specific one, then we'll fail.
			return fmt.Errorf("did not find edge %v", txStr(chanPoint))
		}

		return nil
	}, defaultTimeout)
	require.NoError(t.t, err)

	return policies
}

// waitForNTxsInMempool polls until finding the desired number of transactions
// in the provided miner's mempool. An error is returned if this number is not
// met after the given timeout.
func waitForNTxsInMempool(miner *rpcclient.Client, n int,
	timeout time.Duration) ([]*chainhash.Hash, error) {

	breakTimeout := time.After(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var err error
	var mempool []*chainhash.Hash
	for {
		select {
		case <-breakTimeout:
			return nil, fmt.Errorf("wanted %v, found %v txs "+
				"in mempool: %v", n, len(mempool), mempool)
		case <-ticker.C:
			mempool, err = miner.GetRawMempool()
			if err != nil {
				return nil, err
			}

			if len(mempool) == n {
				return mempool, nil
			}
		}
	}
}

// getPaymentResult reads a final result from the stream and returns it.
func getPaymentResult(stream routerrpc.Router_SendPaymentV2Client) (*lnrpc.Payment, error) {
	for {
		payment, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		if payment.Status != lnrpc.Payment_IN_FLIGHT {
			return payment, nil
		}
	}
}

func assertNumPendingHTLCs(numHTLCs int, nodes ...*lntest.HarnessNode) func() error {
	return func() error {
		ctxb := context.Background()

		listReq := &lnrpc.ListChannelsRequest{}
		pending := 0
		for _, node := range nodes {
			ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			nodeChans, err := node.ListChannels(ctxt, listReq)
			if err != nil {
				return fmt.Errorf("Unable to list channels"+
					"for node %s: %v", node.Name(), err)
			}

			for _, channel := range nodeChans.Channels {
				pending += len(channel.PendingHtlcs)
			}
		}

		if numHTLCs != pending {
			return fmt.Errorf("expected %d HTLCs, got %d", numHTLCs, pending)
		}

		return nil
	}
}

// assertAmountSent generates a closure which queries sndr and rcvr's currently
// active channels and asserts that sndr sent amt satoshis
// and rcvr received amt satoshis.
func assertAmountSent(amt btcutil.Amount, sndr, rcvr *lntest.HarnessNode) func() error {
	return func() error {
		// Both channels should also have properly accounted for the
		// amount that has been sent/received over the channel.
		listChReq := &lnrpc.ListChannelsRequest{}

		ctxb := context.Background()
		sndrChannels := make(map[uint64]bool)

		// List sender channels so as to tally payment amounts over them.
		ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		sndrListChannels, err := sndr.ListChannels(ctxt, listChReq)
		if err != nil {
			return fmt.Errorf("unable to query %s's channel list: %v",
				sndr.Name(), err)
		}
		for _, channel := range sndrListChannels.Channels {
			sndrChannels[channel.ChanId] = true
		}

		var sndrSatoshisSent int64
		listPayReq := &lnrpc.ListPaymentsRequest{}

		ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		// TODO: Assuming all payments are returned with a single request.
		sndrListPayments, err := sndr.ListPayments(ctxt, listPayReq)
		if err != nil {
			return fmt.Errorf("unable  to query %s's payments: %v",
				sndr.Name(), err)
		}
		// Only take into account active sender channels.
		for _, payment := range sndrListPayments.Payments {
			if payment.Status != lnrpc.Payment_SUCCEEDED {
				continue
			}
			for _, htlc := range payment.Htlcs {
				if htlc.Status != lnrpc.HTLCAttempt_SUCCEEDED {
					continue
				}
				firstRouteCh := htlc.Route.Hops[0].ChanId
				if _, ok := sndrChannels[firstRouteCh]; ok {
					sndrSatoshisSent += payment.ValueSat
				}
			}
		}
		if sndrSatoshisSent != int64(amt) {
			return fmt.Errorf("%s's satoshis sent is incorrect "+
				"got %v, expected %v", sndr.Name(),
				sndrSatoshisSent, amt.ToUnit(btcutil.AmountSatoshi))
		}

		ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		rcvrListChannels, err := rcvr.ListChannels(ctxt, listChReq)
		if err != nil {
			return fmt.Errorf("unable to query %s's channel list: %v",
				rcvr.Name(), err)
		}
		var rcvrSatoshisReceived int64
		for _, channel := range rcvrListChannels.Channels {
			rcvrSatoshisReceived += channel.TotalSatoshisReceived
		}
		if rcvrSatoshisReceived != int64(amt) {
			return fmt.Errorf("%s's satoshis sent is incorrect "+
				"got %v, expected %v", rcvr.Name(),
				rcvrSatoshisReceived, amt.ToUnit(btcutil.AmountSatoshi))
		}

		return nil
	}
}

func findChanFundingTx(ctx context.Context, t *harnessTest,
	chanPoint *lnrpc.ChannelPoint, node *lntest.HarnessNode) *lnrpc.Transaction {

	fundingTxID, err := lnrpc.GetChanPointFundingTxid(chanPoint)
	if err != nil {
		t.Fatalf("unable to convert funding txid into chainhash.Hash:"+
			" %v", err)
	}
	target := fundingTxID.String()

	txns, err := node.LightningClient.GetTransactions(
		ctx, &lnrpc.GetTransactionsRequest{})
	if err != nil {
		t.Fatalf("could not get transactions: %v", err)
	}

	for _, tx := range txns.Transactions {
		if tx.TxHash == target {
			return tx
		}
	}

	return nil
}

func createThreeHopNetwork(t *harnessTest, net *lntest.NetworkHarness,
	alice, bob *lntest.HarnessNode, carolHodl bool, c lnrpc.CommitmentType) (
	*lnrpc.ChannelPoint, *lnrpc.ChannelPoint, *lntest.HarnessNode) {

	ctxb := context.Background()

	net.EnsureConnected(t.t, alice, bob)

	// Make sure there are enough utxos for anchoring.
	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, alice)
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, bob)
	}

	// We'll start the test by creating a channel between Alice and Bob,
	// which will act as the first leg for out multi-hop HTLC.
	const chanAmt = 1000000
	var aliceFundingShim *lnrpc.FundingShim
	var thawHeight uint32
	if c == lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE {
		_, minerHeight, err := net.Miner.Client.GetBestBlock()
		require.NoError(t.t, err)
		thawHeight = uint32(minerHeight + 144)
		aliceFundingShim, _, _ = deriveFundingShim(
			net, t, alice, bob, chanAmt, thawHeight, true,
		)
	}
	aliceChanPoint := openChannelAndAssert(
		t, net, alice, bob,
		lntest.OpenChannelParams{
			Amt:            chanAmt,
			CommitmentType: c,
			FundingShim:    aliceFundingShim,
		},
	)

	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err := alice.WaitForNetworkChannelOpen(ctxt, aliceChanPoint)
	if err != nil {
		t.Fatalf("alice didn't report channel: %v", err)
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = bob.WaitForNetworkChannelOpen(ctxt, aliceChanPoint)
	if err != nil {
		t.Fatalf("bob didn't report channel: %v", err)
	}

	// Next, we'll create a new node "carol" and have Bob connect to her. If
	// the carolHodl flag is set, we'll make carol always hold onto the
	// HTLC, this way it'll force Bob to go to chain to resolve the HTLC.
	carolFlags := nodeArgsForCommitType(c)
	if carolHodl {
		carolFlags = append(carolFlags, "--hodl.exit-settle")
	}
	carol := net.NewNode(t.t, "Carol", carolFlags)

	net.ConnectNodes(t.t, bob, carol)

	// Make sure Carol has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, carol)
	}

	// We'll then create a channel from Bob to Carol. After this channel is
	// open, our topology looks like:  A -> B -> C.
	var bobFundingShim *lnrpc.FundingShim
	if c == lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE {
		bobFundingShim, _, _ = deriveFundingShim(
			net, t, bob, carol, chanAmt, thawHeight, true,
		)
	}
	bobChanPoint := openChannelAndAssert(
		t, net, bob, carol,
		lntest.OpenChannelParams{
			Amt:            chanAmt,
			CommitmentType: c,
			FundingShim:    bobFundingShim,
		},
	)
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = bob.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("alice didn't report channel: %v", err)
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = carol.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("bob didn't report channel: %v", err)
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = alice.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("bob didn't report channel: %v", err)
	}

	return aliceChanPoint, bobChanPoint, carol
}

// deriveFundingShim creates a channel funding shim by deriving the necessary
// keys on both sides.
func deriveFundingShim(net *lntest.NetworkHarness, t *harnessTest,
	carol, dave *lntest.HarnessNode, chanSize btcutil.Amount,
	thawHeight uint32, publish bool) (*lnrpc.FundingShim,
	*lnrpc.ChannelPoint, *chainhash.Hash) {

	ctxb := context.Background()
	keyLoc := &walletrpc.KeyReq{KeyFamily: 9999}
	carolFundingKey, err := carol.WalletKitClient.DeriveNextKey(ctxb, keyLoc)
	require.NoError(t.t, err)
	daveFundingKey, err := dave.WalletKitClient.DeriveNextKey(ctxb, keyLoc)
	require.NoError(t.t, err)

	// Now that we have the multi-sig keys for each party, we can manually
	// construct the funding transaction. We'll instruct the backend to
	// immediately create and broadcast a transaction paying out an exact
	// amount. Normally this would reside in the mempool, but we just
	// confirm it now for simplicity.
	_, fundingOutput, err := input.GenFundingPkScript(
		carolFundingKey.RawKeyBytes, daveFundingKey.RawKeyBytes,
		int64(chanSize),
	)
	require.NoError(t.t, err)

	var txid *chainhash.Hash
	targetOutputs := []*wire.TxOut{fundingOutput}
	if publish {
		txid, err = net.Miner.SendOutputsWithoutChange(
			targetOutputs, 5,
		)
		require.NoError(t.t, err)
	} else {
		tx, err := net.Miner.CreateTransaction(targetOutputs, 5, false)
		require.NoError(t.t, err)

		txHash := tx.TxHash()
		txid = &txHash
	}

	// At this point, we can being our external channel funding workflow.
	// We'll start by generating a pending channel ID externally that will
	// be used to track this new funding type.
	var pendingChanID [32]byte
	_, err = rand.Read(pendingChanID[:])
	require.NoError(t.t, err)

	// Now that we have the pending channel ID, Dave (our responder) will
	// register the intent to receive a new channel funding workflow using
	// the pending channel ID.
	chanPoint := &lnrpc.ChannelPoint{
		FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: txid[:],
		},
	}
	chanPointShim := &lnrpc.ChanPointShim{
		Amt:       int64(chanSize),
		ChanPoint: chanPoint,
		LocalKey: &lnrpc.KeyDescriptor{
			RawKeyBytes: daveFundingKey.RawKeyBytes,
			KeyLoc: &lnrpc.KeyLocator{
				KeyFamily: daveFundingKey.KeyLoc.KeyFamily,
				KeyIndex:  daveFundingKey.KeyLoc.KeyIndex,
			},
		},
		RemoteKey:     carolFundingKey.RawKeyBytes,
		PendingChanId: pendingChanID[:],
		ThawHeight:    thawHeight,
	}
	fundingShim := &lnrpc.FundingShim{
		Shim: &lnrpc.FundingShim_ChanPointShim{
			ChanPointShim: chanPointShim,
		},
	}
	_, err = dave.FundingStateStep(ctxb, &lnrpc.FundingTransitionMsg{
		Trigger: &lnrpc.FundingTransitionMsg_ShimRegister{
			ShimRegister: fundingShim,
		},
	})
	require.NoError(t.t, err)

	// If we attempt to register the same shim (has the same pending chan
	// ID), then we should get an error.
	_, err = dave.FundingStateStep(ctxb, &lnrpc.FundingTransitionMsg{
		Trigger: &lnrpc.FundingTransitionMsg_ShimRegister{
			ShimRegister: fundingShim,
		},
	})
	if err == nil {
		t.Fatalf("duplicate pending channel ID funding shim " +
			"registration should trigger an error")
	}

	// We'll take the chan point shim we just registered for Dave (the
	// responder), and swap the local/remote keys before we feed it in as
	// Carol's funding shim as the initiator.
	fundingShim.GetChanPointShim().LocalKey = &lnrpc.KeyDescriptor{
		RawKeyBytes: carolFundingKey.RawKeyBytes,
		KeyLoc: &lnrpc.KeyLocator{
			KeyFamily: carolFundingKey.KeyLoc.KeyFamily,
			KeyIndex:  carolFundingKey.KeyLoc.KeyIndex,
		},
	}
	fundingShim.GetChanPointShim().RemoteKey = daveFundingKey.RawKeyBytes

	return fundingShim, chanPoint, txid
}

// nodeArgsForCommitType returns the command line flag to supply to enable this
// commitment type.
func nodeArgsForCommitType(commitType lnrpc.CommitmentType) []string {
	switch commitType {
	case lnrpc.CommitmentType_LEGACY:
		return []string{"--protocol.legacy.committweak"}
	case lnrpc.CommitmentType_STATIC_REMOTE_KEY:
		return []string{}
	case lnrpc.CommitmentType_ANCHORS:
		return []string{"--protocol.anchors"}
	case lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE:
		return []string{
			"--protocol.anchors",
			"--protocol.script-enforced-lease",
		}
	}

	return nil
}
