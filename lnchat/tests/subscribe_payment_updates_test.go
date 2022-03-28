package itest

import (
	"context"
	"crypto/rand"
	"encoding/json"
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

	"github.com/c13n-io/c13n-go/lnchat"
)

var (
	allPaymentStatuses = []lnrpc.Payment_PaymentStatus{
		lnrpc.Payment_UNKNOWN,
		lnrpc.Payment_IN_FLIGHT,
		lnrpc.Payment_SUCCEEDED,
		lnrpc.Payment_FAILED,
	}
	paymentStatusAssoc = map[lnchat.PaymentStatus]lnrpc.Payment_PaymentStatus{
		lnchat.PaymentUNKNOWN:   lnrpc.Payment_UNKNOWN,
		lnchat.PaymentINFLIGHT:  lnrpc.Payment_IN_FLIGHT,
		lnchat.PaymentSUCCEEDED: lnrpc.Payment_SUCCEEDED,
		lnchat.PaymentFAILED:    lnrpc.Payment_FAILED,
	}
)

func testSubscribePaymentUpdates(net *lntest.NetworkHarness, t *harnessTest) {
	type testcase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testcase{
		{
			name: "stop on context cancel",
			test: testSubscribePaymentUpdatesTermination,
		},
		{
			name: "payment update constraints",
			test: testSubscribePaymentUpdatesConstraints,
		},
	}

	ctxb := context.Background()

	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	// Open 100k satoshi channel.
	chanAmt := btcutil.Amount(100_000)
	chanPoint := openChannelAndAssert(
		t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	// Wait for Alice and Bob to recognize and advertise the channel.
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err := net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
	assert.NoErrorf(t.t, err, "alice didn't "+
		"advertise channel before timeout: %v", err)
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
	assert.NoErrorf(t.t, err, "bob didn't "+
		"advertise channel before timeout: %v", err)

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

	err = wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob),
		pendingHTLCTimeout,
	)
	assert.NoErrorf(t.t, err, "unable to assert no pending htlcs: %v", err)

	// Close the channel.
	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)
}

func testSubscribePaymentUpdatesTermination(
	net *lntest.NetworkHarness, t *harnessTest) {
	// Open a subscription and cancel the context, checking that:
	// a single error update is received for the cancelled context, and
	// the update channel is closed afterwards.

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)
	defer mgrAlice.Close()

	ctxb := context.Background()

	ctxt, cancel := context.WithCancel(ctxb)

	var lastKnownPaymentIdx uint64
	payments := retrievePayments(t.t, net.Alice, true)
	if len(payments) > 0 {
		lastKnownPaymentIdx = payments[len(payments)-1].PaymentIndex
	}

	permissiveFilterFunc := func(_ *lnchat.Payment) bool {
		return true
	}
	updateChan, err := mgrAlice.SubscribePaymentUpdates(ctxt,
		lastKnownPaymentIdx, permissiveFilterFunc)
	assert.NoError(t.t, err, "could not create payment update subcription")

	time.Sleep(500 * time.Millisecond)
	cancel()

	var updates []lnchat.PaymentUpdate
	for update := range updateChan {
		updates = append(updates, update)
	}

	expectedUpdates := []lnchat.PaymentUpdate{
		lnchat.PaymentUpdate{
			Err: context.Canceled,
		},
	}
	assert.Equal(t.t, expectedUpdates, updates)
}

func testSubscribePaymentUpdatesConstraints(
	net *lntest.NetworkHarness, t *harnessTest) {

	testInvoiceIntents := []struct {
		valueMsat int64
		settle    bool
	}{
		{100000, true},
		{210000, true},
		{123456, false},
		{500000, true},
	}

	type testcase struct {
		name                    string
		deletePrevious          bool
		includePreviousPayments bool
		allowedStatuses         []lnrpc.Payment_PaymentStatus
	}

	cases := []testcase{
		{
			name:            "all updates",
			deletePrevious:  true,
			allowedStatuses: allPaymentStatuses,
		},
		{
			name:            "all updates for new payments",
			deletePrevious:  false,
			allowedStatuses: allPaymentStatuses,
		},
		{
			name:                    "all updates with existing payments",
			deletePrevious:          false,
			includePreviousPayments: true,
			allowedStatuses:         allPaymentStatuses,
		},
		{
			name:           "only final updates",
			deletePrevious: true,
			allowedStatuses: []lnrpc.Payment_PaymentStatus{
				lnrpc.Payment_SUCCEEDED,
				lnrpc.Payment_FAILED,
			},
		},
	}

	for _, c := range cases {
		t.t.Run(c.name, func(t1 *testing.T) {
			payer, payee := net.Alice, net.Bob

			if c.deletePrevious {
				deletePaymentsAndAssert(t1, payer)
			}
			var startIndex uint64
			if c.includePreviousPayments {
				payments := retrievePayments(t1, payer, true)
				if len(payments) > 1 {
					last := payments[len(payments)-1]
					startIndex = last.GetPaymentIndex()
				}
			}

			makePayments := func(t *testing.T) []*lnrpc.Payment {
				return preparePaymentHistory(t, payer, payee,
					testInvoiceIntents)
			}

			testSubscribePaymentUpdatesCase(t1, payer,
				startIndex, c.allowedStatuses, makePayments)
		})
	}
}

func testSubscribePaymentUpdatesCase(t *testing.T, node *lntest.HarnessNode,
	startIdx uint64, allowedStatuses []lnrpc.Payment_PaymentStatus,
	facilitatePayments func(*testing.T) []*lnrpc.Payment) {

	var maxTestReceivedUpdates = 100

	mgr, err := createNodeManager(node)
	assert.NoError(t, err)
	defer mgr.Close()

	// Retrieve existing payments after requested index,
	// since their updates will be received from the subscription
	// Note: Only final updates are received for them,
	// so the filter aplies only for index
	preexistingPayments := filterPaymentUpdates(
		retrievePayments(t, node, true), startIdx, allowedStatuses)

	// Create payment update subscription
	subscriptionFilter := func(p *lnchat.Payment) bool {
		// NOTE: Ignore some initial updates
		// that seem to be received sometimes.
		if len(p.Htlcs) == 0 {
			return false
		}

		for _, s := range allowedStatuses {
			if paymentStatusAssoc[p.Status] == s {
				return true
			}
		}
		return false
	}

	// Forward subscription updates to a buffered channel
	bufferSubscription := func(
		in <-chan lnchat.PaymentUpdate) <-chan lnchat.PaymentUpdate {
		ch := make(chan lnchat.PaymentUpdate, maxTestReceivedUpdates)

		go func() {
			defer close(ch)
			for update := range in {
				ch <- update
			}
		}()

		return ch
	}

	ctxb := context.Background()

	ctxt, cancel := context.WithCancel(ctxb)
	paymentUpdateCh, err := mgr.SubscribePaymentUpdates(ctxt,
		startIdx, subscriptionFilter)
	assert.NoError(t, err, "could not create payment update subcription")
	subscriptionChan := bufferSubscription(paymentUpdateCh)
	defer func() {
		// Drain update channel, and assert that only
		// a context cancellation update was remaining.
		expected := []lnchat.PaymentUpdate{
			lnchat.PaymentUpdate{Err: context.Canceled},
		}

		var remaining []lnchat.PaymentUpdate
		for rem := range subscriptionChan {
			remaining = append(remaining, rem)
		}
		assert.Equal(t, expected, remaining)
	}()
	defer cancel()

	// Facilitate payments
	paymentUpdates := facilitatePayments(t)

	// Retrieve subscription updates and check they match the actual ones
	expectedUpdates := append(preexistingPayments, filterPaymentUpdates(
		paymentUpdates, startIdx, allowedStatuses)...)
	waitAndAssertPaymentUpdatesMatch(t, expectedUpdates, subscriptionChan)
}

func waitAndAssertPaymentUpdatesMatch(t *testing.T,
	expectedUpdates []*lnrpc.Payment,
	updateCh <-chan lnchat.PaymentUpdate) {

	checkUpdateMatch := func(update *lnchat.Payment,
		expected *lnrpc.Payment) bool {

		switch true {
		case update == nil || expected == nil:
			return false
		case update.Hash != expected.PaymentHash:
			return false
		case paymentStatusAssoc[update.Status] != expected.Status:
			return false
		case len(update.Htlcs) == len(expected.Htlcs):
			for i, htlc := range update.Htlcs {
				if htlc.Status != expected.Htlcs[i].Status {
					return false
				}
			}
			return true
		}

		return false
	}

	receivedUpdates := make(map[int]*lnchat.Payment)
	for received := range updateCh {
		if received.Err != nil {
			assert.Failf(t, "received unexpected "+
				"payment subscription error",
				"%+v", received.Err)
		}

		found, idx := false, 0
		for i, expected := range expectedUpdates {
			if checkUpdateMatch(received.Payment, expected) {
				found, idx = true, i
				break
			}
		}

		switch true {
		case found && (receivedUpdates[idx] != nil):
			encoded, _ := json.MarshalIndent(received, "", "  ")
			assert.Fail(t, "received duplicate payment update",
				"%s", encoded)
		case found:
			receivedUpdates[idx] = received.Payment
		default:
			encoded, _ := json.MarshalIndent(received, "", "  ")
			assert.Failf(t, "received unexpected payment update",
				"%s", encoded)
		}

		// Terminate once all expected updates are received
		if len(receivedUpdates) == len(expectedUpdates) {
			break
		}
	}
}

// Create a minimal payment history between 2 nodes,
// and return the list of received payment updates.
func preparePaymentHistory(t *testing.T,
	payer, payee *lntest.HarnessNode,
	invoiceIntents []struct {
		valueMsat int64
		settle    bool
	}) []*lnrpc.Payment {

	t.Helper()

	var paymentTimeoutSecs int32 = 40

	type paymentAttempt = struct {
		preimage   lntypes.Preimage
		invUpdates invoicesrpc.Invoices_SubscribeSingleInvoiceClient
		payUpdates routerrpc.Router_SendPaymentV2Client
	}

	attempts := make([]paymentAttempt, len(invoiceIntents))

	ctxb := context.Background()

	// Create invoices and attempt payment
	for i, intent := range invoiceIntents {
		// Create invoice on payee
		attempts[i].preimage = mustNewPreimage(t)

		holdInvoiceReq := &invoicesrpc.AddHoldInvoiceRequest{
			Hash:      getPreimageHash(attempts[i].preimage),
			ValueMsat: intent.valueMsat,
		}

		ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		holdInvoice, err := payee.AddHoldInvoice(ctxt, holdInvoiceReq)
		assert.NoErrorf(t, err, "%s could not create hold invoice",
			payee.Name())

		// Create payee invoice subscription
		invSubReq := &invoicesrpc.SubscribeSingleInvoiceRequest{
			RHash: getPreimageHash(attempts[i].preimage),
		}

		ctxt, cancel = context.WithTimeout(ctxb, 2*defaultTimeout)
		defer cancel()
		invUpdates, err := payee.SubscribeSingleInvoice(ctxt, invSubReq)
		assert.NoErrorf(t, err, "%s could not create invoice subscription",
			payee.Name())
		attempts[i].invUpdates = invUpdates

		// Attempt payment from payer
		payReq := &routerrpc.SendPaymentRequest{
			PaymentRequest: holdInvoice.PaymentRequest,
			TimeoutSeconds: paymentTimeoutSecs,
		}

		ctxt, cancel = context.WithTimeout(ctxb, 2*defaultTimeout)
		defer cancel()
		payUpdates, err := payer.RouterClient.SendPaymentV2(ctxt, payReq)
		assert.NoErrorf(t, err, "%s could not send payment", payer.Name())
		attempts[i].payUpdates = payUpdates

		initialUpdate, err := waitForPaymentStatus(t, payUpdates,
			lnrpc.Payment_IN_FLIGHT)
		assert.NoError(t, err, "%s did not receive "+
			"initial payment update", payer.Name())
		assert.Len(t, initialUpdate, 1)
	}

	var paymentUpdateList []*lnrpc.Payment

	// Resolve invoices and wait for final state updates to be reported
	for i, intent := range invoiceIntents {
		attempt := attempts[i]

		finalInvoiceState := lnrpc.Invoice_SETTLED
		finalPaymentStatus := lnrpc.Payment_SUCCEEDED

		switch intent.settle {
		case true:
			// In case the invoice should be settled,
			// wait for it to be reported as accepted
			// and then settle it.
			_, err := waitForInvoiceState(t,
				attempt.invUpdates, lnrpc.Invoice_ACCEPTED)
			assert.NoErrorf(t, err, "%s did not report invoice "+
				"as accepted", payee.Name())

			settleReq := &invoicesrpc.SettleInvoiceMsg{
				Preimage: attempt.preimage[:],
			}

			ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			_, err = payee.SettleInvoice(ctxt, settleReq)
			assert.NoErrorf(t, err, "%s could not settle invoice",
				payee.Name())
		default:
			// Otherwise, cancel it.
			cancelReq := &invoicesrpc.CancelInvoiceMsg{
				PaymentHash: getPreimageHash(attempt.preimage),
			}

			ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			_, err := payee.CancelInvoice(ctxt, cancelReq)
			assert.NoErrorf(t, err, "%s could not cancel invoice",
				payee.Name())

			finalInvoiceState = lnrpc.Invoice_CANCELED
			finalPaymentStatus = lnrpc.Payment_FAILED
		}

		// Wait for the final state updates to be reported.
		_, err := waitForInvoiceState(t, attempt.invUpdates,
			finalInvoiceState)
		assert.NoErrorf(t, err, "%s did not report invoice as %s",
			payee.Name(), finalInvoiceState)

		updates, err := waitForPaymentStatus(t, attempt.payUpdates,
			finalPaymentStatus)
		assert.NoErrorf(t, err, "%s did not report payment as %s",
			payer.Name(), finalPaymentStatus)

		paymentUpdateList = append(paymentUpdateList, updates...)
	}

	return paymentUpdateList
}

func waitForInvoiceState(t *testing.T,
	invUpdates invoicesrpc.Invoices_SubscribeSingleInvoiceClient,
	state lnrpc.Invoice_InvoiceState) ([]*lnrpc.Invoice, error) {

	t.Helper()

	var updates []*lnrpc.Invoice
	for {
		update, err := invUpdates.Recv()
		if err != nil {
			return updates, err
		}
		updates = append(updates, update)
		if update.State == state {
			break
		}
	}
	return updates, nil
}

func waitForPaymentStatus(t *testing.T,
	payUpdates routerrpc.Router_SendPaymentV2Client,
	status lnrpc.Payment_PaymentStatus) ([]*lnrpc.Payment, error) {

	t.Helper()

	var updates []*lnrpc.Payment
	for {
		update, err := payUpdates.Recv()
		if err != nil {
			return updates, err
		}
		updates = append(updates, update)
		if update.Status == status {
			break
		}
	}
	return updates, nil
}

// Retrieve a node's payment list.
// In case it has none, nil is returned.
func retrievePayments(t *testing.T, node *lntest.HarnessNode,
	includeIncomplete bool) []*lnrpc.Payment {

	ctx := context.Background()
	ctxt, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req := &lnrpc.ListPaymentsRequest{
		IncludeIncomplete: includeIncomplete,
	}
	paymentList, err := node.ListPayments(ctxt, req)
	assert.NoErrorf(t, err, "%s could not retrieve payment list: %v",
		node.Name(), err)

	if len(paymentList.GetPayments()) == 0 {
		return nil
	}
	return paymentList.GetPayments()
}

// Return only the updates after the specified index
// and whose status matches one of the provided ones.
func filterPaymentUpdates(updates []*lnrpc.Payment, afterIdx uint64,
	permittedStatuses []lnrpc.Payment_PaymentStatus) []*lnrpc.Payment {

	var filtered []*lnrpc.Payment
	for _, update := range updates {
		if update.PaymentIndex <= afterIdx {
			continue
		}
		for _, status := range permittedStatuses {
			if update.Status == status {
				filtered = append(filtered, update)
				break
			}
		}
	}
	return filtered
}

// Delete a node's payment history,
// and verify that node reports no payments.
func deletePaymentsAndAssert(t *testing.T, node *lntest.HarnessNode) {
	ctxb := context.Background()

	// Delete payer's payment history
	deleteReq := &lnrpc.DeleteAllPaymentsRequest{}

	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	_, err := node.DeleteAllPayments(ctxt, deleteReq)
	assert.NoErrorf(t, err, "%s could not delete payments", node.Name())

	// Assert that payer reports no payments
	listReq := &lnrpc.ListPaymentsRequest{
		IncludeIncomplete: true,
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	paymentList, err := node.ListPayments(ctxt, listReq)
	assert.NoErrorf(t, err, "%s could not retrieve payment list", node.Name())
	assert.Emptyf(t, paymentList.Payments,
		"%s reported non-empty payment list", node.Name())
}

func mustNewPreimage(t *testing.T) lntypes.Preimage {
	t.Helper()

	var preimage lntypes.Preimage
	_, err := rand.Read(preimage[:])
	assert.NoError(t, err, "could not generate preimage")

	return preimage
}

func getPreimageHash(preimage lntypes.Preimage) []byte {
	h := preimage.Hash()
	return h[:]
}
