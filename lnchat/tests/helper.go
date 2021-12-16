package itest

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
		h.lndHarness.SaveProfilesPages()
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
func openChannelStream(ctx context.Context, t *harnessTest,
	net *lntest.NetworkHarness, alice, bob *lntest.HarnessNode,
	p lntest.OpenChannelParams) lnrpc.Lightning_OpenChannelClient {

	t.t.Helper()

	// Wait until we are able to fund a channel successfully. This wait
	// prevents us from erroring out when trying to create a channel while
	// the node is starting up.
	var chanOpenUpdate lnrpc.Lightning_OpenChannelClient
	err := wait.NoError(func() error {
		var err error
		chanOpenUpdate, err = net.OpenChannel(ctx, alice, bob, p)
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
func openChannelAndAssert(ctx context.Context, t *harnessTest,
	net *lntest.NetworkHarness, alice, bob *lntest.HarnessNode,
	p lntest.OpenChannelParams) *lnrpc.ChannelPoint {

	t.t.Helper()

	chanOpenUpdate := openChannelStream(ctx, t, net, alice, bob, p)

	// Mine 6 blocks, then wait for Alice's node to notify us that the
	// channel has been opened. The funding transaction should be found
	// within the first newly mined block. We mine 6 blocks so that in the
	// case that the channel is public, it is announced to the network.
	block := mineBlocks(t, net, 6, 1)[0]

	fundingChanPoint, err := net.WaitForChannelOpen(ctx, chanOpenUpdate)
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
	if err := net.AssertChannelExists(ctx, alice, &chanPoint); err != nil {
		t.Fatalf("unable to assert channel existence: %v", err)
	}
	if err := net.AssertChannelExists(ctx, bob, &chanPoint); err != nil {
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
func closeChannelAndAssert(ctx context.Context, t *harnessTest,
	net *lntest.NetworkHarness, node *lntest.HarnessNode,
	fundingChanPoint *lnrpc.ChannelPoint, force bool) *chainhash.Hash {

	return closeChannelAndAssertType(ctx, t, net, node, fundingChanPoint, false, force)
}

func closeChannelAndAssertType(ctx context.Context, t *harnessTest,
	net *lntest.NetworkHarness, node *lntest.HarnessNode,
	fundingChanPoint *lnrpc.ChannelPoint, anchors, force bool) *chainhash.Hash {

	// Fetch the current channel policy. If the channel is currently
	// enabled, we will register for graph notifications before closing to
	// assert that the node sends out a disabling update as a result of the
	// channel being closed.
	curPolicy := getChannelPolicies(t, node, node.PubKeyStr, fundingChanPoint)[0]
	expectDisable := !curPolicy.Disabled

	// If the current channel policy is enabled, begin subscribing the graph
	// updates before initiating the channel closure.
	var graphSub *graphSubscription
	if expectDisable {
		sub := subscribeGraphNotifications(t, ctx, node)
		graphSub = &sub
		defer close(graphSub.quit)
	}

	closeUpdates, _, err := net.CloseChannel(ctx, node, fundingChanPoint, force)
	if err != nil {
		t.Fatalf("unable to close channel: %v", err)
	}

	// If the channel policy was enabled prior to the closure, wait until we
	// received the disabled update.
	if expectDisable {
		curPolicy.Disabled = true
		waitForChannelUpdate(
			t, *graphSub,
			[]expectedChanUpdate{
				{node.PubKeyStr, curPolicy, fundingChanPoint},
			},
		)
	}

	return assertChannelClosed(
		ctx, t, net, node, fundingChanPoint, anchors, closeUpdates,
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
	ctxt, _ := context.WithTimeout(ctx, defaultTimeout)
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
	ctxt, _ = context.WithTimeout(ctx, defaultTimeout)
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

	closingTxid, err := net.WaitForChannelClose(ctx, closeUpdates)
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

// commitType is a simple enum used to run though the basic funding flow with
// different commitment formats.
type commitType byte

const (
	// commitTypeLegacy is the old school commitment type.
	commitTypeLegacy commitType = iota

	// commiTypeTweakless is the commitment type where the remote key is
	// static (non-tweaked).
	commitTypeTweakless

	// commitTypeAnchors is the kind of commitment that has extra outputs
	// used for anchoring down to commitment using CPFP.
	commitTypeAnchors
)

// String returns that name of the commitment type.
func (c commitType) String() string {
	switch c {
	case commitTypeLegacy:
		return "legacy"
	case commitTypeTweakless:
		return "tweakless"
	case commitTypeAnchors:
		return "anchors"
	default:
		return "invalid"
	}
}

// Args returns the command line flag to supply to enable this commitment type.
func (c commitType) Args() []string {
	switch c {
	case commitTypeLegacy:
		return []string{"--protocol.committweak"}
	case commitTypeTweakless:
		return []string{}
	case commitTypeAnchors:
		return []string{"--protocol.anchors"}
	}

	return nil
}

// calcStaticFee calculates appropriate fees for commitment transactions.  This
// function provides a simple way to allow test balance assertions to take fee
// calculations into account.
func (c commitType) calcStaticFee(numHTLCs int) btcutil.Amount {
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
	if c == commitTypeAnchors {
		feePerKw = chainfee.SatPerKVByte(
			lnwallet.DefaultAnchorsCommitMaxFeeRateSatPerVByte * 1000,
		).FeePerKWeight()
		commitWeight = input.AnchorCommitWeight
		anchors = 2 * anchorSize
	}

	return feePerKw.FeeForWeight(int64(commitWeight+htlcWeight*numHTLCs)) +
		anchors
}

// channelCommitType retrieves the active channel commitment type for the given
// chan point.
func channelCommitType(node *lntest.HarnessNode,
	chanPoint *lnrpc.ChannelPoint) (commitType, error) {

	ctxb := context.Background()
	ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)

	req := &lnrpc.ListChannelsRequest{}
	channels, err := node.ListChannels(ctxt, req)
	if err != nil {
		return 0, fmt.Errorf("listchannels failed: %v", err)
	}

	for _, c := range channels.Channels {
		if c.ChannelPoint == txStr(chanPoint) {
			switch c.CommitmentType {

			// If the anchor output size is non-zero, we are
			// dealing with the anchor type.
			case lnrpc.CommitmentType_ANCHORS:
				return commitTypeAnchors, nil

			// StaticRemoteKey means it is tweakless,
			case lnrpc.CommitmentType_STATIC_REMOTE_KEY:
				return commitTypeTweakless, nil

			// Otherwise legacy.
			default:
				return commitTypeLegacy, nil
			}
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

// expectedChanUpdate houses params we expect a ChannelUpdate to advertise.
type expectedChanUpdate struct {
	advertisingNode string
	expectedPolicy  *lnrpc.RoutingPolicy
	chanPoint       *lnrpc.ChannelPoint
}

// waitForChannelUpdate waits for a node to receive the expected channel
// updates.
func waitForChannelUpdate(t *harnessTest, subscription graphSubscription,
	expUpdates []expectedChanUpdate) {

	// Create an array indicating which expected channel updates we have
	// received.
	found := make([]bool, len(expUpdates))
out:
	for {
		select {
		case graphUpdate := <-subscription.updateChan:
			for _, update := range graphUpdate.ChannelUpdates {
				if len(expUpdates) == 0 {
					t.Fatalf("received unexpected channel "+
						"update from %v for channel %v",
						update.AdvertisingNode,
						update.ChanId)
				}

				// For each expected update, check if it matches
				// the update we just received.
				for i, exp := range expUpdates {
					fundingTxStr := txStr(update.ChanPoint)
					if fundingTxStr != txStr(exp.chanPoint) {
						continue
					}

					if update.AdvertisingNode !=
						exp.advertisingNode {
						continue
					}

					err := checkChannelPolicy(
						update.RoutingPolicy,
						exp.expectedPolicy,
					)
					if err != nil {
						continue
					}

					// We got a policy update that matched
					// the values and channel point of what
					// we expected, mark it as found.
					found[i] = true

					// If we have no more channel updates
					// we are waiting for, break out of the
					// loop.
					rem := 0
					for _, f := range found {
						if !f {
							rem++
						}
					}

					if rem == 0 {
						break out
					}

					// Since we found a match among the
					// expected updates, break out of the
					// inner loop.
					break
				}
			}
		case err := <-subscription.errChan:
			t.Fatalf("unable to recv graph update: %v", err)
		case <-time.After(defaultTimeout):
			if len(expUpdates) == 0 {
				return
			}
			t.Fatalf("did not receive channel update")
		}
	}
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
	ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)
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

// checkChannelPolicy checks that the policy matches the expected one.
func checkChannelPolicy(policy, expectedPolicy *lnrpc.RoutingPolicy) error {
	if policy.FeeBaseMsat != expectedPolicy.FeeBaseMsat {
		return fmt.Errorf("expected base fee %v, got %v",
			expectedPolicy.FeeBaseMsat, policy.FeeBaseMsat)
	}
	if policy.FeeRateMilliMsat != expectedPolicy.FeeRateMilliMsat {
		return fmt.Errorf("expected fee rate %v, got %v",
			expectedPolicy.FeeRateMilliMsat,
			policy.FeeRateMilliMsat)
	}
	if policy.TimeLockDelta != expectedPolicy.TimeLockDelta {
		return fmt.Errorf("expected time lock delta %v, got %v",
			expectedPolicy.TimeLockDelta,
			policy.TimeLockDelta)
	}
	if policy.MinHtlc != expectedPolicy.MinHtlc {
		return fmt.Errorf("expected min htlc %v, got %v",
			expectedPolicy.MinHtlc, policy.MinHtlc)
	}
	if policy.MaxHtlcMsat != expectedPolicy.MaxHtlcMsat {
		return fmt.Errorf("expected max htlc %v, got %v",
			expectedPolicy.MaxHtlcMsat, policy.MaxHtlcMsat)
	}
	if policy.Disabled != expectedPolicy.Disabled {
		return errors.New("edge should be disabled but isn't")
	}

	return nil
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

// graphSubscription houses the proxied update and error chans for a node's
// graph subscriptions.
type graphSubscription struct {
	updateChan chan *lnrpc.GraphTopologyUpdate
	errChan    chan error
	quit       chan struct{}
}

//revive:disable:context-as-argument

// subscribeGraphNotifications subscribes to channel graph updates and launches
// a goroutine that forwards these to the returned channel.
func subscribeGraphNotifications(t *harnessTest, ctxb context.Context,
	node *lntest.HarnessNode) graphSubscription {

	// We'll first start by establishing a notification client which will
	// send us notifications upon detected changes in the channel graph.
	req := &lnrpc.GraphTopologySubscription{}
	ctx, cancelFunc := context.WithCancel(ctxb)
	topologyClient, err := node.SubscribeChannelGraph(ctx, req)
	if err != nil {
		t.Fatalf("unable to create topology client: %v", err)
	}

	// We'll launch a goroutine that will be responsible for proxying all
	// notifications recv'd from the client into the channel below.
	errChan := make(chan error, 1)
	quit := make(chan struct{})
	graphUpdates := make(chan *lnrpc.GraphTopologyUpdate, 20)
	go func() {
		for {
			defer cancelFunc()

			select {
			case <-quit:
				return
			default:
				graphUpdate, err := topologyClient.Recv()
				select {
				case <-quit:
					return
				default:
				}

				if err == io.EOF {
					return
				} else if err != nil {
					select {
					case errChan <- err:
					case <-quit:
					}
					return
				}

				select {
				case graphUpdates <- graphUpdate:
				case <-quit:
					return
				}
			}
		}
	}()

	return graphSubscription{
		updateChan: graphUpdates,
		errChan:    errChan,
		quit:       quit,
	}
}

//revive:enable:context-as-argument

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
			ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)
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
		ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)
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

		ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
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

		ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
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
	alice, bob *lntest.HarnessNode, carolHodl bool, c commitType) (
	*lnrpc.ChannelPoint, *lnrpc.ChannelPoint, *lntest.HarnessNode) {

	ctxb := context.Background()

	ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)
	net.EnsureConnected(ctxt, t.t, alice, bob)

	// Make sure there are enough utxos for anchoring.
	for i := 0; i < 2; i++ {
		ctxt, _ = context.WithTimeout(context.Background(), defaultTimeout)
		net.SendCoins(ctxt, t.t, btcutil.SatoshiPerBitcoin, alice)

		ctxt, _ = context.WithTimeout(context.Background(), defaultTimeout)
		net.SendCoins(ctxt, t.t, btcutil.SatoshiPerBitcoin, bob)
	}

	// We'll start the test by creating a channel between Alice and Bob,
	// which will act as the first leg for our multi-hop HTLC.
	const chanAmt = 1000000
	ctxt, _ = context.WithTimeout(ctxb, channelOpenTimeout)
	aliceChanPoint := openChannelAndAssert(
		ctxt, t, net, alice, bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	err := alice.WaitForNetworkChannelOpen(ctxt, aliceChanPoint)
	if err != nil {
		t.Fatalf("alice didn't report channel: %v", err)
	}

	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	err = bob.WaitForNetworkChannelOpen(ctxt, aliceChanPoint)
	if err != nil {
		t.Fatalf("bob didn't report channel: %v", err)
	}

	// Next, we'll create a new node "carol" and have Bob connect to her. If
	// the carolHodl flag is set, we'll make carol always hold onto the
	// HTLC, this way it'll force Bob to go to chain to resolve the HTLC.
	carolFlags := c.Args()
	if carolHodl {
		carolFlags = append(carolFlags, "--hodl.exit-settle")
	}
	carol := net.NewNode(t.t, "Carol", carolFlags)
	if err != nil {
		t.Fatalf("unable to create new node: %v", err)
	}
	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	net.ConnectNodes(ctxt, t.t, bob, carol)

	// Make sure Carol has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		ctxt, _ = context.WithTimeout(context.Background(), defaultTimeout)
		net.SendCoins(ctxt, t.t, btcutil.SatoshiPerBitcoin, carol)
	}

	// We'll then create a channel from Bob to Carol. After this channel is
	// open, our topology looks like:  A -> B -> C.
	ctxt, _ = context.WithTimeout(ctxb, channelOpenTimeout)
	bobChanPoint := openChannelAndAssert(
		ctxt, t, net, bob, carol,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)
	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	err = bob.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("bob didn't report channel: %v", err)
	}
	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	err = carol.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("carol didn't report channel: %v", err)
	}
	ctxt, _ = context.WithTimeout(ctxb, defaultTimeout)
	err = alice.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
	if err != nil {
		t.Fatalf("alice didn't report channel: %v", err)
	}

	return aliceChanPoint, bobChanPoint, carol
}
