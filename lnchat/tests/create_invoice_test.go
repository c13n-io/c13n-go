package itest

import (
	"context"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testCreateInvoice(net *lntest.NetworkHarness, t *harnessTest) {
	cases := []struct {
		name         string
		memo         string
		amt          lnchat.Amount
		expiry       int64
		privateHints bool
		expectedErr  error
	}{
		{
			name:   "normal",
			memo:   "first invoice",
			amt:    lnchat.NewAmount(10235),
			expiry: 3456,
		},
		{
			name:   "no memo",
			amt:    lnchat.NewAmount(10000),
			expiry: 3000,
		},
	}

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	for _, c := range cases {
		t.t.Run(c.name, func(tt *testing.T) {
			ctxb := context.Background()

			inv, err := mgrAlice.CreateInvoice(ctxb,
				c.memo, c.amt, c.expiry, c.privateHints)
			assert.NoError(t.t, err)

			switch c.expectedErr {
			default:
				assert.EqualError(t.t, err, c.expectedErr.Error())
			case nil:
				assert.Equal(t.t, inv.Memo, c.memo)
				assert.EqualValues(t.t, inv.Value, c.amt.Msat())
				assert.Equal(t.t, inv.Expiry, c.expiry)
				assert.Equal(t.t, inv.Private, c.privateHints)

				// Invoice cleanup
				hash, err := lntypes.MakeHashFromStr(inv.Hash)
				assert.NoError(t.t, err)

				cancelReq := &invoicesrpc.CancelInvoiceMsg{
					PaymentHash: hash[:],
				}
				_, err = net.Alice.CancelInvoice(ctxb, cancelReq)
				assert.NoError(t.t, err)
			}
		})
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
