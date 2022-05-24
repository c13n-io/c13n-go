package itest

import (
	"context"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/record"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testGetRoute(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "Single Hop",
			test: testGetRouteSingleHop,
		},
		{
			name: "Multi Hop",
			test: testGetRouteMultiHop,
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

func testGetRouteSingleHop(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, alice, bob *lntest.HarnessNode)
	}

	subTests := []testCase{
		{
			name: "Immediate Neighbour, No fees",
			test: testGetRouteSingleHopNoFees,
		},
	}

	// Make sure Alice has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	// Open a channel with 100k satoshis between Alice and Bob with Alice being
	// the sole funder of the channel.
	chanAmt := btcutil.Amount(1000000)
	chanPoint := openChannelAndAssert(
		t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	// Wait for Alice and Bob to recognize and advertise the new channel
	// generated above.
	err := net.Alice.WaitForNetworkChannelOpen(chanPoint)
	if err != nil {
		t.Fatalf("alice didn't advertise channel before "+
			"timeout: %v", err)
	}

	err = net.Bob.WaitForNetworkChannelOpen(chanPoint)
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

func testGetRouteSingleHopNoFees(t *harnessTest, alice, bob *lntest.HarnessNode) {
	mgrAlice, err := createNodeManager(alice)
	assert.NoError(t.t, err)

	var recordTypeKey uint64 = record.CustomTypeStart + 311

	// Alice creates the message
	recipient := bob.PubKeyStr
	amount := lnchat.NewAmount(1000)
	payOpts := lnchat.PaymentOptions{
		FeeLimitMsat:   0,
		FinalCltvDelta: 60,
	}

	payload := map[uint64][]byte{
		recordTypeKey: []byte("test"),
	}

	expectedResponse := lnchat.Route{
		Amt:  lnchat.NewAmount(1000),
		Fees: lnchat.NewAmount(0),
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, "", payOpts, payload)

	assert.Equal(t.t, expectedResponse.Amt, response.Amt)
	assert.Equal(t.t, expectedResponse.Fees, response.Fees)
	assert.Len(t.t, response.Hops, 1)
	assert.NoError(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetRouteMultiHop(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, source, dest *lntest.HarnessNode)
	}

	subTests := []testCase{
		{
			name: "Intermediate, with fees",
			test: testGetRouteMultiHopWithFees,
		},
		{
			name: "No route found",
			test: testGetRouteMultiHopNoRouteFound,
		},
		{
			name: "with payment request",
			test: testGetRoutePayReq,
		},
	}

	aliceBobChanPoint, bobCarolChanPoint, carol := createThreeHopNetwork(t,
		net, net.Alice, net.Bob, false,
		lnrpc.CommitmentType_STATIC_REMOTE_KEY,
	)

	// Create an additional private channel between Bob and Carol.
	const privChanAmt = 400000
	privBobCarolChanPoint := openChannelAndAssert(t,
		net, net.Bob, carol,
		lntest.OpenChannelParams{
			Amt:     privChanAmt,
			PushAmt: privChanAmt >> 1,
			Private: true,
		},
	)

	err := net.Bob.WaitForNetworkChannelOpen(privBobCarolChanPoint)
	if err != nil {
		t.Fatalf("bob didn't advertise channel before timeout: %v", err)
	}

	err = carol.WaitForNetworkChannelOpen(privBobCarolChanPoint)
	if err != nil {
		t.Fatalf("carol didn't advertise channel before timeout: %v", err)
	}

	for _, subTest := range subTests {
		// Needed in case of parallel testing.
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(ht, net.Alice, carol)
		})

		if !success {
			break
		}
	}

	if err := wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob, carol),
		pendingHTLCTimeout,
	); err != nil {
		t.Fatalf("Unable to assert no pending htlcs: %v", err)
	}

	// Close Alice -> Bob channel.
	closeChannelAndAssert(t, net, net.Alice, aliceBobChanPoint, false)
	// Close Bob -> Carol channel.
	closeChannelAndAssert(t, net, net.Bob, bobCarolChanPoint, false)
	// Close private Bob -> Carol channel.
	closeChannelAndAssert(t, net, net.Bob, privBobCarolChanPoint, false)

	shutdownAndAssert(net, t, carol)
}

func testGetRouteMultiHopWithFees(t *harnessTest, alice, dest *lntest.HarnessNode) {
	mgrAlice, err := createNodeManager(alice)
	assert.NoError(t.t, err)

	var recordTypeKey uint64 = record.CustomTypeStart + 311

	// Alice creates the message
	recipient := dest.PubKeyStr
	amount := lnchat.NewAmount(1000)
	payOpts := lnchat.PaymentOptions{
		FeeLimitMsat:   1000,
		FinalCltvDelta: 60,
	}

	payload := map[uint64][]byte{
		recordTypeKey: []byte("test"),
	}

	expectedResponse := lnchat.Route{
		Amt:  lnchat.NewAmount(1000),
		Fees: lnchat.NewAmount(1000),
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, "", payOpts, payload)

	assert.Equal(t.t, expectedResponse.Amt, response.Amt)
	assert.Equal(t.t, expectedResponse.Fees, response.Fees)
	assert.Len(t.t, response.Hops, 2)
	assert.NoError(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetRouteMultiHopNoRouteFound(t *harnessTest, alice, dest *lntest.HarnessNode) {
	mgrAlice, err := createNodeManager(alice)
	assert.NoError(t.t, err)

	var recordTypeKey uint64 = record.CustomTypeStart + 311

	// Alice creates the message
	recipient := dest.PubKeyStr
	amount := lnchat.NewAmount(1000)
	payOpts := lnchat.PaymentOptions{
		FeeLimitMsat:   10,
		FinalCltvDelta: 60,
	}

	payload := map[uint64][]byte{
		recordTypeKey: []byte("test"),
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, "", payOpts, payload)

	assert.Nil(t.t, response)
	assert.Error(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testGetRoutePayReq(t *harnessTest, alice, dest *lntest.HarnessNode) {
	type testParams struct {
		name        string
		privInvoice bool
	}

	subTests := []testParams{
		{
			name:        "no route hints",
			privInvoice: false,
		},
		{
			name:        "with route hints",
			privInvoice: true,
		},
	}

	const invoiceAmt = 20000

	mgrAlice, err := createNodeManager(alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	for _, subTest := range subTests {
		t.t.Run(subTest.name, func(t1 *testing.T) {
			// Destination node creates a payment request
			invoiceReq := &lnrpc.Invoice{
				Memo:    "test payReq hints:" + subTest.name,
				Value:   invoiceAmt,
				Private: subTest.privInvoice,
			}

			ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			invoiceResp, err := dest.AddInvoice(ctxt, invoiceReq)
			assert.NoError(t.t, err)
			assert.NotNil(t.t, invoiceResp)

			// Verify route hint existence on payment request
			ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			destInvoice, err := dest.LookupInvoice(ctxt,
				&lnrpc.PaymentHash{
					RHash: invoiceResp.RHash,
				},
			)
			assert.NoError(t.t, err)
			switch subTest.privInvoice {
			case true:
				assert.NotEmpty(t.t, destInvoice.RouteHints)
			default:
				assert.Empty(t.t, destInvoice.RouteHints)
			}

			// Source attempts route discovery by providing the invoice
			payReq := invoiceResp.GetPaymentRequest()
			payOpts := lnchat.PaymentOptions{
				FinalCltvDelta: 60,
			}

			ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			routeResp, _, err := mgrAlice.GetRoute(ctxt,
				"", lnchat.NewAmount(0), payReq, payOpts, nil)

			assert.NoError(t.t, err)
			assert.NotEmpty(t.t, routeResp)

			// Destination cancels the invoice
			cancelReq := &invoicesrpc.CancelInvoiceMsg{
				PaymentHash: invoiceResp.GetRHash(),
			}

			ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
			defer cancel()
			_, err = dest.CancelInvoice(ctxt, cancelReq)
			assert.NoError(t.t, err)
		})
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
