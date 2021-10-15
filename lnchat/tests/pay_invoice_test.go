package itest

import (
	"context"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/record"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-backend/lnchat"
)

func testSendPaymentSubscribeInvoiceUpdates(net *lntest.NetworkHarness, t *harnessTest) {
	// A TLV record key following the "it's okay to be odd" rule.
	var recordTypeKey uint64 = record.CustomTypeStart + 311

	cases := []struct {
		name          string
		recipient     string
		amt           lnchat.Amount
		invoiceAmt    int64
		withPayReq    bool
		payload       map[uint64][]byte
		expectedErr   error
		paymentStatus lnchat.PaymentStatus
	}{
		{
			name:       "Keysend with payload",
			recipient:  net.Bob.PubKeyStr,
			amt:        lnchat.NewAmount(1000),
			invoiceAmt: 1000,
			withPayReq: false,
			payload: map[uint64][]byte{
				recordTypeKey: []byte("test"),
			},
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:          "Payment to invoice",
			recipient:     "",
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    0,
			withPayReq:    true,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
	}

	// Make sure Alice has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.

	for i := 0; i < 2; i++ {
		ctxt, _ := context.WithTimeout(context.Background(), defaultTimeout)
		net.SendCoins(ctxt, t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	// Open a channel with 1M satoshis between Alice and Bob with Alice being
	// the sole funder of the channel.
	ctxt, _ := context.WithTimeout(context.Background(), channelOpenTimeout)
	chanAmt := btcutil.Amount(1000000)
	chanPoint := openChannelAndAssert(
		ctxt, t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	// Wait for Alice and Bob to recognize and advertise the new channel
	// generated above.
	ctxt, _ = context.WithTimeout(context.Background(), defaultTimeout)
	err := net.Alice.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("alice didn't advertise channel before "+
			"timeout: %v", err)
	}
	ctxt, _ = context.WithTimeout(context.Background(), defaultTimeout)
	err = net.Bob.WaitForNetworkChannelOpen(ctxt, chanPoint)
	if err != nil {
		t.Fatalf("bob didn't advertise channel before "+
			"timeout: %v", err)
	}

	// Create managers
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	mgrBob, err := createNodeManager(net.Bob)
	assert.NoError(t.t, err)

	paymentFilter := func(p *lnchat.Payment) bool {
		return p.Status == lnchat.PaymentSUCCEEDED ||
			p.Status == lnchat.PaymentFAILED
	}
	invoiceFilter := func(inv *lnchat.Invoice) bool {
		return inv.State == lnchat.InvoiceSETTLED
	}

	for _, c := range cases {
		t.t.Run(c.name, func(subTest *testing.T) {
			ctxb := context.Background()

			// Setup invoice update channel
			ctxc, cancel := context.WithCancel(ctxb)
			defer cancel()
			invSubscription, err := mgrBob.SubscribeInvoiceUpdates(ctxc,
				0, invoiceFilter)

			assert.NotNil(t.t, invSubscription)
			assert.NoError(t.t, err, "Failed to create invoice subscription")

			// If a payReq is required
			var payReq string
			switch c.withPayReq {
			case true:
				// Bob generates a new invoice
				invoice := &lnrpc.Invoice{
					ValueMsat: c.invoiceAmt,
				}

				ctxt, _ := context.WithTimeout(ctxb, defaultTimeout)
				invoiceResp, err := net.Bob.AddInvoice(ctxt, invoice)
				assert.NoError(t.t, err, "Invoice generation failed")
				payReq = invoiceResp.PaymentRequest
			default:
				payReq = ""
			}

			paymentUpdates, err := mgrAlice.SendPayment(ctxb,
				c.recipient, c.amt, payReq,
				lnchat.PaymentOptions{TimeoutSecs: 30},
				c.payload, paymentFilter)

			switch c.expectedErr {
			case nil:
				assert.NoError(t.t, err)

				// Check payment update
				paymentUpdate := <-paymentUpdates
				payment, err := paymentUpdate.Payment, paymentUpdate.Err
				assert.NoError(t.t, err, "Payment failed")
				assert.Equal(t.t, c.paymentStatus, payment.Status)

				// Check subscription update
				invUpdate := <-invSubscription
				inv, err := invUpdate.Inv, invUpdate.Err
				assert.NoError(t.t, err, "Invoice update failed")
				assert.NotNil(t.t, inv)
				assert.Equal(t.t, lnchat.InvoiceSETTLED, inv.State)

				// Check that payment details match
				switch c.withPayReq {
				case true:
					assert.EqualValues(t.t,
						payment.PaymentRequest,
						inv.PaymentRequest)
				default:
					assert.EqualValues(t.t,
						payment.Hash, inv.Hash)
				}

				if c.payload != nil {
					payloadRecords := inv.GetCustomRecords()[0]

					assert.Len(t.t, payloadRecords, 2)
					assert.Contains(t.t, payloadRecords, recordTypeKey)
					assert.Equal(t.t, c.payload[recordTypeKey],
						payloadRecords[recordTypeKey])
				}
			default:
				assert.Error(t.t, err)
			}

		})
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)

	err = mgrBob.Close()
	assert.NoError(t.t, err)

	if err := wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob),
		pendingHTLCTimeout,
	); err != nil {
		t.Fatalf("Unable to assert no pending htlcs: %v", err)
	}

	// Close the channel.
	ctxt, _ = context.WithTimeout(context.Background(), channelCloseTimeout)
	closeChannelAndAssert(ctxt, t, net, net.Alice, chanPoint, false)
}
