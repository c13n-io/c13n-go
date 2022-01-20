package itest

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stateWaitTimeout = 30 * time.Second

func amountSats(amtSats int64) btcutil.Amount {
	return btcutil.Amount(amtSats)
}

func satToMsat(sat int64) uint64 {
	return uint64(sat * 1000)
}

func testGetSelfBalance(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "wallet balance (confirmed)",
			test: testGetSelfBalanceConfirmed,
		},
		{
			name: "wallet balance (unconfirmed)",
			test: testGetSelfBalanceUnconfirmed,
		},
		{
			name: "channel balance (open channel)",
			test: testGetSelfBalanceChannel,
		},
		{
			name: "pending open channel",
			test: testGetSelfBalancePendingCh,
		},
		{
			name: "unsettled balance on channel",
			test: testGetSelfBalanceUnsettled,
		},
	}

	for _, subTest := range subTests {
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(net, ht)
		})

		if !success {
			break
		}
	}
}

func testGetSelfBalanceConfirmed(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	balance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	var amt int64 = btcutil.SatoshiPerBitcoin
	net.SendCoins(t.t, amountSats(amt), net.Alice)

	updatedBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// The confirmed wallet balance should update.
	expectedBalance := balance
	expectedBalance.WalletConfirmedBalanceSat += amt
	assert.EqualValues(t.t, expectedBalance, updatedBalance)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetSelfBalanceUnconfirmed(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	balance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	var amt int64 = btcutil.SatoshiPerBitcoin
	net.SendCoinsUnconfirmed(t.t, amountSats(amt), net.Alice)

	updatedBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// The unconfirmed wallet balance should update.
	expectedBalance := balance
	expectedBalance.WalletUnconfirmedBalanceSat += amt
	assert.EqualValues(t.t, expectedBalance, updatedBalance)

	// Confirm the balance
	_, err = net.Miner.Client.Generate(6)
	assert.NoError(t.t, err)
	newConfirmedBalance := balance.WalletConfirmedBalanceSat + amt
	err = net.Alice.WaitForBalance(amountSats(newConfirmedBalance), true)
	assert.NoError(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetSelfBalanceChannel(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	balance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	var amt, pushAmt int64 = 100000, 1000

	chanPoint := openChannelAndAssert(t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt:     amountSats(amt),
			PushAmt: amountSats(pushAmt),
		},
	)

	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()

	err = net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("alice didn't advertise channel before "+
			"timeout: %v", err)
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()

	err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("bob didn't advertise channel before "+
			"timeout: %v", err)
	}

	updatedBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// The total channel balance should update to:
	// * The available channel balance will be the previous one
	// plus the balance of the new channel
	// minus the push amount and commit fees.
	// * The total confirmed balance will be the previous minus
	// the amount used to open the channel (including push amount)
	// as well as the fees used for the transaction.
	cType, err := channelCommitType(net.Alice, chanPoint)
	require.NoError(t.t, err, "unable to get channel type")

	commitFee := int64(calcStaticFee(cType, 0))

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	tx := findChanFundingTx(ctxt, t, chanPoint, net.Alice)
	txFees := tx.TotalFees

	expectedBalance := balance
	expectedBalance.WalletConfirmedBalanceSat -= amt + txFees
	expectedBalance.ChannelBalance.LocalMsat += satToMsat(
		amt - pushAmt - commitFee)
	expectedBalance.ChannelBalance.RemoteMsat += satToMsat(pushAmt)
	assert.EqualValues(t.t, expectedBalance, updatedBalance)

	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetSelfBalancePendingCh(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	// Retrieve balance information previous to channel open.
	balance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// Open a channel in pending state.
	var amt, pushAmt int64 = 100000, 1000

	pendingUpdate, err := net.OpenPendingChannel(
		net.Alice, net.Bob, amountSats(amt), amountSats(pushAmt))
	if err != nil {
		t.Fatalf("unable to open channel: %v", err)
	}

	// Retrieve the channel and transaction information.
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	pendingChs, err := net.Alice.PendingChannels(ctxt,
		&lnrpc.PendingChannelsRequest{},
	)
	assert.NoError(t.t, err)
	assert.Len(t.t, pendingChs.PendingOpenChannels, 1)

	chanPoint := &lnrpc.ChannelPoint{
		FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: pendingUpdate.Txid,
		},
		OutputIndex: pendingUpdate.OutputIndex,
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	tx := findChanFundingTx(ctxt, t, chanPoint, net.Alice)

	// Retrieve balance information following channel open.
	updatedBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// The total channel balance should update to:
	// * The total confirmed balance will be the previous minus
	// the amount used to open the channel as well as the transaction
	// fees and the new unconfirmed balance.
	// * The pending channel balance will be the previous plus
	// the amount used to open the channel minus the commit fees.
	newUnconfirmed := updatedBalance.WalletUnconfirmedBalanceSat
	txFees := tx.TotalFees
	commitFee := pendingChs.PendingOpenChannels[0].CommitFee

	expectedBalance := balance
	expectedBalance.WalletUnconfirmedBalanceSat = newUnconfirmed
	expectedBalance.WalletConfirmedBalanceSat -= newUnconfirmed +
		amt + txFees
	expectedBalance.PendingOpenBalance.LocalMsat += satToMsat(
		amt - pushAmt - commitFee)
	expectedBalance.PendingOpenBalance.RemoteMsat += satToMsat(pushAmt)
	assert.EqualValues(t.t, expectedBalance, updatedBalance)

	// Perform channel cleanup.
	mineBlocks(t, net, 6, 1)

	// Ensure that the channel is open.
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("channel not seen on network before timeout:"+
			" %v", err)
	}

	// Close channel.
	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetSelfBalanceUnsettled(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)
	mgrBob, err := createNodeManager(net.Bob)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	var amt, pushAmt int64 = 100000, 1000

	chanPoint := openChannelAndAssert(t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt:     amountSats(amt),
			PushAmt: amountSats(pushAmt),
		},
	)

	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("alice didn't advertise channel before "+
			"timeout: %v", err)
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("bob didn't advertise channel before "+
			"timeout: %v", err)
	}

	aliceBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)
	bobBalance, err := mgrBob.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	var billedAmtMsat uint64 = 20357
	// Bob creates a hold invoice.
	var preimage lntypes.Preimage
	if _, err = rand.Read(preimage[:]); err != nil {
		t.Fatalf("unable to generate invoice preimage: %v", err)
	}
	preimageHash := preimage.Hash()
	holdInvoiceReq := &invoicesrpc.AddHoldInvoiceRequest{
		Hash:      preimageHash[:],
		ValueMsat: int64(billedAmtMsat),
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	holdInvResp, err := net.Bob.AddHoldInvoice(ctxt, holdInvoiceReq)
	if err != nil {
		t.Fatalf("unable to add hold invoice: %v", err)
	}
	// Bob subscribes to the invoice and waits until it is OPEN.
	invSubscribeReq := &invoicesrpc.SubscribeSingleInvoiceRequest{
		RHash: preimageHash[:],
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceStream, err := net.Bob.SubscribeSingleInvoice(ctxt, invSubscribeReq)
	if err != nil {
		t.Fatalf("unable to subscribe to invoice: %v", err)
	}
	if err = wait.NoError(func() error {
		invoice, err := invoiceStream.Recv()
		if err != nil {
			return err
		}
		if invoice.State != lnrpc.Invoice_OPEN {
			return fmt.Errorf("expected invoice state OPEN, got %v",
				invoice.State)
		}
		return nil
	}, stateWaitTimeout); err != nil {
		t.Fatalf("could not verify invoice as OPEN: %v", err)
	}
	// Alice pays the invoice.
	sendPayReq := &routerrpc.SendPaymentRequest{
		PaymentRequest: holdInvResp.PaymentRequest,
		TimeoutSeconds: 60,
	}
	ctxt, cancel = context.WithCancel(ctxb)
	defer cancel()
	paymentStream, err := net.Alice.RouterClient.SendPaymentV2(
		ctxt, sendPayReq)

	if err != nil {
		t.Fatalf("unable to pay hold invoice: %v", err)
	}
	// Bob waits until the invoice is ACCEPTED.
	if err = wait.NoError(func() error {
		invoice, err := invoiceStream.Recv()
		if err != nil {
			return err
		}
		if invoice.State != lnrpc.Invoice_ACCEPTED {
			return fmt.Errorf("expected invoice state ACCEPTED, got %v",
				invoice.State)
		}
		return nil
	}, stateWaitTimeout); err != nil {
		t.Fatalf("could not verify invoice as ACCEPTED: %v", err)
	}

	updatedAliceBalance, err := mgrAlice.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)
	updatedBobBalance, err := mgrBob.GetSelfBalance(ctxb)
	assert.NoError(t.t, err)

	// The total channel balance should update to:
	// * The available channel balance will be the previous one
	// minus the paid amount.
	// * The remote unsettled channel balance will be
	// the previous one plus the sent amount.
	expectedAliceBalance := aliceBalance
	expectedAliceBalance.ChannelBalance.LocalMsat -= billedAmtMsat
	expectedAliceBalance.UnsettledBalance.RemoteMsat += billedAmtMsat
	assert.EqualValues(t.t, expectedAliceBalance, updatedAliceBalance)

	expectedBobBalance := bobBalance
	expectedBobBalance.ChannelBalance.RemoteMsat -= billedAmtMsat
	expectedBobBalance.UnsettledBalance.LocalMsat += billedAmtMsat
	assert.EqualValues(t.t, expectedBobBalance, updatedBobBalance)

	// Settle the invoice before cleanup.
	settleInvReq := &invoicesrpc.SettleInvoiceMsg{
		Preimage: preimage[:],
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	if _, err := net.Bob.SettleInvoice(ctxt, settleInvReq); err != nil {
		t.Fatalf("unable to settle hold invoice: %v", err)
	}
	// Verify that Bob sees the invoice as paid.
	if err = wait.NoError(func() error {
		invoice, err := invoiceStream.Recv()
		if err != nil {
			return err
		}
		if invoice.State != lnrpc.Invoice_SETTLED {
			return fmt.Errorf("expected invoice state SETTLED, got %v",
				invoice.State)
		}
		return nil
	}, stateWaitTimeout); err != nil {
		t.Fatalf("could not verify invoice as SETTLED: %v", err)
	}
	// Verify that Alice's payment was successful.
	payResult, err := func(stream routerrpc.Router_SendPaymentV2Client) (
		*lnrpc.Payment, error) {

		for {
			payment, err := stream.Recv()
			if err != nil {
				return nil, err
			}
			if payment.Status != lnrpc.Payment_IN_FLIGHT {
				return payment, nil
			}
		}
	}(paymentStream)
	if err != nil {
		t.Fatalf("failed to retrieve payment result: %v", err)
	}
	assert.EqualValues(t.t, payResult.Status, lnrpc.Payment_SUCCEEDED)

	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
	err = mgrBob.Close()
	assert.NoError(t.t, err)
}
