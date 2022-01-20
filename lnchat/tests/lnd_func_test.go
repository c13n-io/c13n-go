package itest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/lntypes"
)

func testSingleHopTests(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, alice, bob *lntest.HarnessNode)
	}

	singleHopSubTests := []testCase{
		{
			name: "invoice generation and payment (AddInvoice-SendPaymentSync)",
			test: testInvoicePaymentSync,
		},
		{
			name: "invoice generation and payment (AddInvoice-SendPaymentV2)",
			test: testInvoicePaymentV2,
		},
	}

	for _, subTest := range singleHopSubTests {
		// Needed in case of parallel testing.
		subTest := subTest

		ctxb := context.Background()

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
		ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		err := net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
		if err != nil {
			t.Fatalf("alice didn't advertise channel before "+
				"timeout: %v", err)
		}
		err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
		if err != nil {
			t.Fatalf("bob didn't advertise channel before "+
				"timeout: %v", err)
		}

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(ht, net.Alice, net.Bob)
		})

		if !success {
			break
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
}

func testMultiHopTests(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(t *harnessTest, alice, bob *lntest.HarnessNode)
	}

	multiHopSubTests := []testCase{
		{
			name: "invoice generation and payment (AddInvoice-SendPaymentSync)",
			test: testInvoicePaymentSync,
		},
		{
			name: "invoice generation and payment (AddInvoice-SendPaymentV2)",
			test: testInvoicePaymentV2,
		},
	}

	// We will create a new node carol, and have bob connect to her.
	carol := net.NewNode(t.t, "Carol", nil)

	net.ConnectNodes(t.t, net.Bob, carol)

	for _, subTest := range multiHopSubTests {
		// Needed in case of parallel testing.
		subTest := subTest

		ctxb := context.Background()

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
		ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		err := net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
		if err != nil {
			t.Fatalf("alice didn't advertise channel before "+
				"timeout: %v", err)
		}
		err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
		if err != nil {
			t.Fatalf("bob didn't advertise channel before "+
				"timeout: %v", err)
		}

		// Open a channel from Bob to Carol.
		// After this, the topology will be: A -> B -> C
		bobChanPoint := openChannelAndAssert(
			t, net, net.Bob, carol,
			lntest.OpenChannelParams{
				Amt: chanAmt,
			},
		)

		ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		err = net.Bob.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
		if err != nil {
			t.Fatalf("bob didn't advertise channel before "+
				"timeout: %v", err)
		}
		ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		err = carol.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
		if err != nil {
			t.Fatalf("carol didn't advertise channel before "+
				"timeout: %v", err)
		}
		ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		err = net.Alice.WaitForNetworkChannelOpen(ctxt, bobChanPoint)
		if err != nil {
			t.Fatalf("alice didn't report channel: %v", err)
		}

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(ht, net.Alice, carol)
		})

		if !success {
			break
		}

		if err := wait.NoError(
			assertNumPendingHTLCs(0, net.Alice, net.Bob, carol),
			pendingHTLCTimeout,
		); err != nil {
			t.Fatalf("Unable to assert no pending htlcs: %v", err)
		}

		// Close the channel.
		closeChannelAndAssert(t, net, net.Alice, chanPoint, false)
		closeChannelAndAssert(t, net, net.Bob, bobChanPoint, false)
	}
}

func testInvoicePaymentSync(t *harnessTest, alice, bob *lntest.HarnessNode) {
	ctxb := context.Background()

	// Now that the channel is open, create an invoice for Bob which
	// expects a payment of 1000 satoshis from Alice paid via a particular
	// preimage.
	const paymentAmt = 1000
	invoice := &lnrpc.Invoice{
		Memo:  "testing",
		Value: paymentAmt,
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceResp, err := bob.AddInvoice(ctxt, invoice)
	if err != nil {
		t.Fatalf("unable to add invoice: %v", err)
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceStream, err := bob.SubscribeSingleInvoice(ctxt,
		&invoicesrpc.SubscribeSingleInvoiceRequest{
			RHash: invoiceResp.RHash,
		},
	)
	if err != nil {
		t.Fatalf("unable to subscribe to invoice: %v", err)
	}
	invoiceUpdate, err := invoiceStream.Recv()
	if err != nil {
		t.Fatalf("failed receiving status update: %v", err)
	}
	if invoiceUpdate.State != lnrpc.Invoice_OPEN {
		t.Fatalf("expected invoice state OPEN, got %v",
			invoiceUpdate.State)
	}
	preimage := invoiceUpdate.RPreimage

	// With the invoice for Bob added, send a payment towards Alice paying
	// to the above generated invoice.
	sendReq := &lnrpc.SendRequest{
		PaymentRequest: invoiceResp.PaymentRequest,
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	resp, err := alice.SendPaymentSync(ctxt, sendReq)
	if err != nil {
		t.Fatalf("unable to send payment: %v", err)
	}

	// Ensure we obtain the proper preimage in the response.
	if resp.PaymentError != "" {
		t.Fatalf("error when attempting recv: %v", resp.PaymentError)
	} else if !bytes.Equal(preimage, resp.PaymentPreimage) {
		t.Fatalf("preimage mismatch: expected %v, got %v", preimage,
			resp.GetPaymentPreimage())
	}

	// Bob's invoice should now be found and marked as settled.
	payHash := &lnrpc.PaymentHash{
		RHash: invoiceResp.RHash,
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	dbInvoice, err := bob.LookupInvoice(ctxt, payHash)
	if err != nil {
		t.Fatalf("unable to lookup invoice: %v", err)
	}

	//nolint:staticcheck // lnd integration test code
	if !dbInvoice.Settled {
		t.Fatalf("bob's invoice should be marked as settled: %v",
			spew.Sdump(dbInvoice))
	}

	// With the payment completed all balance related stats should be
	// properly updated.
	err = wait.NoError(
		assertAmountSent(paymentAmt, alice, bob),
		3*time.Second,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testInvoicePaymentV2(t *harnessTest, alice, bob *lntest.HarnessNode) {
	ctxb := context.Background()

	// Now that the channel is open, create an invoice for Bob which
	// expects a payment of 1000 satoshis from Alice paid via a particular
	// preimage.
	const paymentAmt = 1000
	invoice := &lnrpc.Invoice{
		Memo:  "testing",
		Value: paymentAmt,
	}
	ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceResp, err := bob.AddInvoice(ctxt, invoice)
	if err != nil {
		t.Fatalf("unable to add invoice: %v", err)
	}

	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	invoiceStream, err := bob.InvoicesClient.SubscribeSingleInvoice(ctxt,
		&invoicesrpc.SubscribeSingleInvoiceRequest{
			RHash: invoiceResp.RHash,
		},
	)
	if err != nil {
		t.Fatalf("unable to subscribe to invoice: %v", err)
	}
	invoiceUpdate, err := invoiceStream.Recv()
	if err != nil {
		t.Fatalf("failed receiving status update: %v", err)
	}
	if invoiceUpdate.State != lnrpc.Invoice_OPEN {
		t.Fatalf("expected invoice state OPEN, got %v",
			invoiceUpdate.State)
	}
	preimage := invoiceUpdate.RPreimage

	// With the invoice for Bob added, send a payment to Alice paying
	// to the above generated invoice.
	sendReq := &routerrpc.SendPaymentRequest{
		PaymentRequest: invoiceResp.PaymentRequest,
		TimeoutSeconds: 400,
		FeeLimitSat:    1000000,
		//		NoInflightUpdates: true,
	}
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
	defer cancel()
	paymentUpdates, err := alice.RouterClient.SendPaymentV2(ctxt, sendReq)
	if err != nil {
		t.Fatalf("unable to send payment: %v", err)
	}

	// Check the invoice status.
	for {
		invoiceUpdate, err := invoiceStream.Recv()
		if err != nil {
			t.Fatalf("failed receiving status update: %v", err)
		}
		if invoiceUpdate.State != lnrpc.Invoice_SETTLED {
			t.Fatalf("expected invoice state SETTLED, got %v",
				invoiceUpdate.State)
		} else {
			break
		}
	}

	paymentState, err := getPaymentResult(paymentUpdates)
	if err != nil {
		t.Fatalf("failed receiving status update: %v", err)
	}
	if paymentState.Status != lnrpc.Payment_SUCCEEDED {
		t.Fatalf("payment status unsuccessful: %v\n"+
			"for payment: %v\n",
			paymentState.Status, spew.Sdump(paymentState))
	}

	// Ensure we obtain the proper preimage in the response.
	issuedPreimage, _ := lntypes.MakePreimage(preimage)
	receivedPreimage, _ := lntypes.MakePreimageFromStr(paymentState.PaymentPreimage)
	if issuedPreimage != receivedPreimage {
		t.Fatalf("preimage mismatch: expected %v, got %v", preimage,
			paymentState.PaymentPreimage)
	}

	// With the payment completed all balance related stats should be
	// properly updated.
	err = wait.NoError(
		assertAmountSent(paymentAmt, alice, bob),
		3*time.Second,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
