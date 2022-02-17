package itest

import (
	"context"
	"encoding/hex"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testLookupInvoice(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	ctxb := context.Background()

	// Lookup invoice issued by self
	req := &lnrpc.Invoice{
		ValueMsat: 25000000,
		Memo:      "satoshi",
	}
	resp, _ := net.Alice.AddInvoice(ctxb, req)
	invSet, _ := mgrAlice.LookupInvoice(ctxb, hex.EncodeToString(resp.GetRHash()))
	assert.Equal(t.t, invSet.State, lnchat.InvoiceOPEN)
	assert.EqualValues(t.t, invSet.Value, 25000000)
	assert.EqualValues(t.t, invSet.Memo, "satoshi")

	// Lookup invoice issued by other
	resp, _ = net.Bob.AddInvoice(ctxb, req)
	_, err = mgrAlice.LookupInvoice(ctxb, hex.EncodeToString(resp.GetRHash()))
	assert.Error(t.t, err)
}
