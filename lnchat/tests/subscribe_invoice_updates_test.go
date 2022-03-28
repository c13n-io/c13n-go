package itest

import (
	"context"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/record"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testSubscribeInvoiceUpdates(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "Created",
			test: testSubscribeInvoiceUpdatesCreated,
		},
		{
			name: "Settled",
			test: testSubscribeInvoiceUpdatesSettled,
		},
	}

	for _, subTest := range subTests {
		// Needed in case of parallel testing.
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

func testSubscribeInvoiceUpdatesCreated(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, alice, bob *lntest.HarnessNode)
	}

	subTests := []testCase{
		{
			name: "Invoice Creation",
			test: testSubscribeInvoiceUpdatesCreatedSuccess,
		},
	}

	// Make sure Alice has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	for _, subTest := range subTests {
		// Needed in case of parallel testing.
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(ht, net.Alice, net.Bob)
		})

		if !success {
			break
		}
	}

	if err := wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob),
		pendingHTLCTimeout,
	); err != nil {
		t.Fatalf("Unable to assert no pending htlcs: %v", err)
	}
}

// Test SubscribeInvoiceUpdates for successful invoice creation.
func testSubscribeInvoiceUpdatesCreatedSuccess(t *harnessTest, alice, _ *lntest.HarnessNode) {
	mgrAlice, err := createNodeManager(alice)
	assert.NoError(t.t, err)

	// Setup invoice update channel
	invoiceFilter := func(inv *lnchat.Invoice) bool {
		return inv.State == lnchat.InvoiceOPEN
	}

	ctxb := context.Background()

	ctxc, cancelSubscription := context.WithCancel(ctxb)
	invSubscription, err := mgrAlice.SubscribeInvoiceUpdates(ctxc,
		0, invoiceFilter)

	assert.NotNil(t.t, invSubscription)
	assert.NoError(t.t, err, "Failed to create invoice subscription")

	// Create a new invoice
	const requestedAmtMsat = 1000
	invoice := &lnrpc.Invoice{
		Memo:      "invoice created update test",
		ValueMsat: requestedAmtMsat,
	}
	ctxt, cancel := context.WithTimeout(ctxc, defaultTimeout)
	defer cancel()
	invoiceResp, err := alice.AddInvoice(ctxt, invoice)

	assert.NotNil(t.t, invoiceResp)
	assert.NoError(t.t, err)

	// Check subscription update
	invUpdate := <-invSubscription
	inv, err := invUpdate.Inv, invUpdate.Err
	assert.NoError(t.t, err, "Invoice update failed")
	assert.NotNil(t.t, inv)

	// Cancel the invoice subscription and
	// clear any remaining updates from the update channel.
	cancelSubscription()
	for upd := range invSubscription {
		_ = upd
	}

	assert.Equal(t.t, lnchat.InvoiceOPEN, inv.State)
	assert.Equal(t.t, int64(requestedAmtMsat), inv.Value.Msat())
	assert.Equal(t.t, invoiceResp.GetPaymentRequest(), inv.PaymentRequest)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testSubscribeInvoiceUpdatesSettled(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, alice, bob *lntest.HarnessNode)
	}

	subTests := []testCase{
		{
			name: "No payload",
			test: testSubscribeInvoiceUpdatesSettledNoPayload,
		},
		{
			name: "No payload, no amount",
			test: testSubscribeInvoiceUpdatesSettledNoPayloadNoAmt,
		},
		{
			name: "With payload",
			test: testSubscribeInvoiceUpdatesSettledWithPayload,
		},
	}

	ctxb := context.Background()

	// Make sure Alice has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	// Open a channel with 100k satoshis between Alice and Bob with Alice being
	// the sole funder of the channel.
	chanAmt := btcutil.Amount(100_000)
	chanPoint := openChannelAndAssert(
		t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	// Wait for Alice and Bob to recognize and advertise the new channel
	// generated above.
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	err := net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
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

	for _, subTest := range subTests {
		// Needed in case of parallel testing.
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(ht, net.Alice, net.Bob)
		})

		if !success {
			break
		}
	}

	if err := wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob),
		pendingHTLCTimeout,
	); err != nil {
		t.Fatalf("Unable to assert no pending htlcs: %v", err)
	}

	// Close the channel.
	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)
}

// Test SubscribeInvoiceUpdates for successful invoice settlement
// for an invoice with specified amount when paid to without payload
// (simple case).
func testSubscribeInvoiceUpdatesSettledNoPayload(t *harnessTest, alice, bob *lntest.HarnessNode) {
	mgrBob, err := createNodeManager(bob)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	// Setup invoice update channel
	invoiceFilter := func(inv *lnchat.Invoice) bool {
		return inv.State == lnchat.InvoiceSETTLED
	}

	ctxc, cancel := context.WithCancel(ctxb)
	defer cancel()
	invSubscription, err := mgrBob.SubscribeInvoiceUpdates(ctxc,
		0, invoiceFilter)

	assert.NotNil(t.t, invSubscription)
	assert.NoError(t.t, err, "Failed to create invoice subscription")

	// Bob generates a new invoice
	const requestedAmtMsat = 1000
	invoice := &lnrpc.Invoice{
		Memo:      "invoice settled no pa test",
		ValueMsat: requestedAmtMsat,
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceResp, err := bob.AddInvoice(ctxt, invoice)

	assert.NotNil(t.t, invoiceResp)
	assert.NoError(t.t, err)

	// Alice pays the invoice
	sendPaymentReq := &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoiceResp.GetPaymentRequest(),
		NoInflightUpdates: true,
		TimeoutSeconds:    30,
	}

	paymentStream, err := alice.RouterClient.SendPaymentV2(ctxb, sendPaymentReq)
	assert.NoError(t.t, err, "Invoice payment failed")
	p, err := paymentStream.Recv()
	assert.NoError(t.t, err, "payment update failed")
	assert.Equal(t.t, lnrpc.Payment_SUCCEEDED, p.GetStatus())

	// Check subscription update
	invUpdate := <-invSubscription
	inv, err := invUpdate.Inv, invUpdate.Err
	assert.NoError(t.t, err, "Invoice update failed")
	assert.NotNil(t.t, inv)

	assert.Equal(t.t, lnchat.InvoiceSETTLED, inv.State)
	assert.Equal(t.t, int64(requestedAmtMsat), inv.Value.Msat())
	assert.Equal(t.t, invoiceResp.PaymentRequest, inv.PaymentRequest)

	err = mgrBob.Close()
	assert.NoError(t.t, err)
}

// Test SubscribeInvoiceUpdates for successful invoice settlement
// for an invoice without specified amount when paid to without payload
// (one-off donation case).
func testSubscribeInvoiceUpdatesSettledNoPayloadNoAmt(t *harnessTest, alice, bob *lntest.HarnessNode) {
	mgrBob, err := createNodeManager(bob)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	// Setup invoice update channel
	invoiceFilter := func(inv *lnchat.Invoice) bool {
		return inv.State == lnchat.InvoiceSETTLED
	}

	ctxc, cancel := context.WithCancel(ctxb)
	defer cancel()
	invSubscription, err := mgrBob.SubscribeInvoiceUpdates(ctxc,
		0, invoiceFilter)

	assert.NotNil(t.t, invSubscription)
	assert.NoError(t.t, err, "Failed to create invoice subscription")

	// Bob generates a new invoice
	invoice := &lnrpc.Invoice{
		Memo: "invoice created update test",
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceResp, err := bob.AddInvoice(ctxt, invoice)

	assert.NotNil(t.t, invoiceResp)
	assert.NoError(t.t, err)

	// Alice pays the invoice
	const sentAmtMsat = 1000
	sendPaymentReq := &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoiceResp.GetPaymentRequest(),
		NoInflightUpdates: true,
		TimeoutSeconds:    30,
		AmtMsat:           sentAmtMsat,
	}

	paymentStream, err := alice.RouterClient.SendPaymentV2(ctxb, sendPaymentReq)
	assert.NoError(t.t, err, "Invoice payment failed")
	p, err := paymentStream.Recv()
	assert.NoError(t.t, err, "payment update failed")
	assert.Equal(t.t, lnrpc.Payment_SUCCEEDED, p.GetStatus())

	// Check subscription update
	invUpdate := <-invSubscription
	inv, err := invUpdate.Inv, invUpdate.Err
	assert.NoError(t.t, err, "Invoice update failed")
	assert.NotNil(t.t, inv)

	assert.Equal(t.t, lnchat.InvoiceSETTLED, inv.State)
	assert.Equal(t.t, int64(0), inv.Value.Msat())
	assert.Equal(t.t, int64(sentAmtMsat), inv.AmtPaid.Msat())
	assert.Equal(t.t, invoiceResp.PaymentRequest, inv.PaymentRequest)

	err = mgrBob.Close()
	assert.NoError(t.t, err)
}

// Test SubscribeInvoiceUpdates for successful invoice settlement
// for an invoice with specified amount when paid to with payload.
func testSubscribeInvoiceUpdatesSettledWithPayload(t *harnessTest, alice, bob *lntest.HarnessNode) {
	mgrBob, err := createNodeManager(bob)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	// Setup invoice update channel
	invoiceFilter := func(inv *lnchat.Invoice) bool {
		return inv.State == lnchat.InvoiceSETTLED
	}

	ctxc, cancel := context.WithCancel(ctxb)
	defer cancel()
	invSubscription, err := mgrBob.SubscribeInvoiceUpdates(ctxc,
		0, invoiceFilter)

	assert.NotNil(t.t, invSubscription)
	assert.NoError(t.t, err, "Failed to create invoice subscription")

	// Bob generates a new invoice
	const requestedAmtMsat = 1000
	invoice := &lnrpc.Invoice{
		Memo:      "invoice created update test",
		ValueMsat: requestedAmtMsat,
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceResp, err := bob.AddInvoice(ctxt, invoice)

	assert.NotNil(t.t, invoiceResp)
	assert.NoError(t.t, err)

	// Alice pays the invoice
	var recordTypeKey uint64 = record.CustomTypeStart + 311
	customRecords := map[uint64][]byte{
		recordTypeKey: []byte("test"),
	}

	sendPaymentReq := &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoiceResp.GetPaymentRequest(),
		NoInflightUpdates: true,
		TimeoutSeconds:    30,
		DestCustomRecords: customRecords,
	}

	paymentStream, err := alice.RouterClient.SendPaymentV2(ctxb, sendPaymentReq)
	assert.NoError(t.t, err, "Invoice payment failed")
	p, err := paymentStream.Recv()
	assert.NoError(t.t, err, "payment update failed")
	assert.Equal(t.t, lnrpc.Payment_SUCCEEDED, p.GetStatus())

	// Check subscription update
	invUpdate := <-invSubscription
	inv, err := invUpdate.Inv, invUpdate.Err
	assert.NoError(t.t, err, "Invoice update failed")
	assert.NotNil(t.t, inv)

	assert.Equal(t.t, lnchat.InvoiceSETTLED, inv.State)
	assert.Equal(t.t, int64(requestedAmtMsat), inv.Value.Msat())
	assert.Equal(t.t, invoiceResp.PaymentRequest, inv.PaymentRequest)
	assert.Len(t.t, inv.Htlcs, 1)
	assert.Len(t.t, inv.GetCustomRecords(), 1)
	assert.Equal(t.t, customRecords, inv.GetCustomRecords()[0])

	err = mgrBob.Close()
	assert.NoError(t.t, err)
}
