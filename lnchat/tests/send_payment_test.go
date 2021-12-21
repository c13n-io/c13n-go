package itest

import (
	"context"
	"errors"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/lightningnetwork/lnd/record"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testSendPayment(net *lntest.NetworkHarness, t *harnessTest) {
	var recordTypeKey uint64 = record.CustomTypeStart + 311

	invalidRecipient := "000000000000000000000000000000000000000000000000000000000000000000"

	cases := []struct {
		name          string
		recipient     string
		amt           lnchat.Amount
		invoiceAmt    int64
		payReq        bool
		payload       map[uint64][]byte
		expectedErr   error
		paymentStatus lnchat.PaymentStatus
	}{
		{
			name:          "Recipient not matching payReq destination",
			recipient:     invalidRecipient,
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    1000,
			payReq:        true,
			payload:       nil,
			expectedErr:   errors.New(""),
			paymentStatus: lnchat.PaymentFAILED,
		},
		{
			name:          "Recipient matching payReq destination",
			recipient:     net.Bob.PubKeyStr,
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    1000,
			payReq:        true,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:          "Payment amount not matching payReq amount",
			recipient:     "",
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    500,
			payReq:        true,
			payload:       nil,
			expectedErr:   errors.New(""),
			paymentStatus: lnchat.PaymentFAILED,
		},
		{
			name:          "Payment amount matching payReq amount",
			recipient:     "",
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    1000,
			payReq:        true,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:          "Payment amount 0 with defined payReq amount",
			recipient:     "",
			amt:           lnchat.NewAmount(0),
			invoiceAmt:    1000,
			payReq:        true,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:          "No payload without payReq",
			recipient:     net.Bob.PubKeyStr,
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    1000,
			payReq:        false,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:       "Payload without payReq",
			recipient:  net.Bob.PubKeyStr,
			amt:        lnchat.NewAmount(1000),
			invoiceAmt: 1000,
			payReq:     false,
			payload: map[uint64][]byte{
				recordTypeKey: []byte("test"),
			},
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:       "Payload with payReq",
			recipient:  "",
			amt:        lnchat.NewAmount(1000),
			invoiceAmt: 0,
			payReq:     true,
			payload: map[uint64][]byte{
				recordTypeKey: []byte("test"),
			},
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
		{
			name:          "No payload with payReq",
			recipient:     "",
			amt:           lnchat.NewAmount(1000),
			invoiceAmt:    0,
			payReq:        true,
			payload:       nil,
			expectedErr:   nil,
			paymentStatus: lnchat.PaymentSUCCEEDED,
		},
	}

	ctxb := context.Background()

	// Make sure Alice has enough utxos for anchoring. Because the anchor by
	// itself often doesn't meet the dust limit, a utxo from the wallet
	// needs to be attached as an additional input. This can still lead to a
	// positively-yielding transaction.
	for i := 0; i < 2; i++ {
		ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
		defer cancel()
		net.SendCoins(ctxt, t.t, btcutil.SatoshiPerBitcoin, net.Alice)
	}

	// Open a channel with 1M satoshis between Alice and Bob with Alice being
	// the sole funder of the channel.
	ctxt, cancel := context.WithTimeout(ctxb, channelOpenTimeout)
	defer cancel()
	chanAmt := btcutil.Amount(1000000)
	chanPoint := openChannelAndAssert(
		ctxt, t, net, net.Alice, net.Bob,
		lntest.OpenChannelParams{
			Amt: chanAmt,
		},
	)

	// Wait for Alice and Bob to recognize and advertise the new channel
	// generated above.
	ctxt, cancel = context.WithTimeout(ctxb, defaultTimeout)
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

	// Create managers
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	mgrBob, err := createNodeManager(net.Bob)
	assert.NoError(t.t, err)

	defaultPaymentFilter := func(p *lnchat.Payment) bool {
		return p.Status == lnchat.PaymentSUCCEEDED ||
			p.Status == lnchat.PaymentFAILED
	}

	for _, c := range cases {
		t.t.Run(c.name, func(subTest *testing.T) {
			ctxb := context.Background()

			// If a payReq is required
			var payReq string
			if c.payReq {
				// Bob generates a new invoice
				invoice := &lnrpc.Invoice{
					ValueMsat: c.invoiceAmt,
				}

				ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
				defer cancel()

				invoiceResp, err := net.Bob.AddInvoice(ctxt, invoice)
				assert.NoError(t.t, err, "Invoice generation failed")
				payReq = invoiceResp.PaymentRequest
			} else {
				payReq = ""
			}

			paymentUpdates, err := mgrAlice.SendPayment(ctxb,
				c.recipient, c.amt, payReq,
				lnchat.PaymentOptions{TimeoutSecs: 30},
				c.payload, defaultPaymentFilter)

			switch c.expectedErr {
			case nil:
				assert.NoError(t.t, err)
				paymentUpdate := <-paymentUpdates
				payment, err := paymentUpdate.Payment, paymentUpdate.Err
				assert.NoError(t.t, err)
				assert.Equal(t.t, c.paymentStatus, payment.Status)

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
	ctxt, cancel = context.WithTimeout(ctxb, channelCloseTimeout)
	defer cancel()
	closeChannelAndAssert(ctxt, t, net, net.Alice, chanPoint, false)
}
