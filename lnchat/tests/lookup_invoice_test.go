package itest

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testLookupInvoice(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "Own Invoice",
			test: testOwnLookupInvoice,
		},
		{
			name: "Other Invoice",
			test: testUnknownLookupInvoice,
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

func testOwnLookupInvoice(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	valueMsat := 25000000
	memo := "satoshi"

	ctxb := context.Background()

	// Lookup invoice issued by self
	req := &lnrpc.Invoice{
		ValueMsat: int64(valueMsat),
		Memo:      memo,
	}
	resp, _ := net.Alice.AddInvoice(ctxb, req)
	invSet, _ := mgrAlice.LookupInvoice(ctxb, hex.EncodeToString(resp.GetRHash()))
	assert.Equal(t.t, invSet.State, lnchat.InvoiceOPEN)
	assert.EqualValues(t.t, invSet.Value, valueMsat)
	assert.EqualValues(t.t, invSet.Memo, memo)
}

func testUnknownLookupInvoice(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	valueMsat := 25000000
	memo := "satoshi"

	ctxb := context.Background()

	// Lookup invoice issued by other
	req := &lnrpc.Invoice{
		ValueMsat: int64(valueMsat),
		Memo:      memo,
	}
	resp, _ := net.Bob.AddInvoice(ctxb, req)
	_, err = mgrAlice.LookupInvoice(ctxb, hex.EncodeToString(resp.GetRHash()))
	assert.Error(t.t, err)
}
