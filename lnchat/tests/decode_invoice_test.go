package itest

import (
	"bytes"
	"context"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testDecodePayReq(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest, nodeMgr lnchat.LightManager, node *lntest.HarnessNode)
	}

	subTests := []testCase{
		{
			name: "Specified amount",
			test: func(net *lntest.NetworkHarness, t *harnessTest,
				mgr lnchat.LightManager, node *lntest.HarnessNode) {

				ctxb := context.Background()

				invoice := &lnrpc.Invoice{
					Memo:      "Test",
					RPreimage: bytes.Repeat([]byte("ABCD"), 8),
					ValueMsat: 12341234,
				}
				invoiceResp, err := node.AddInvoice(ctxb, invoice)
				if err != nil {
					t.Fatalf("unable to add invoice: %v", err)
				}

				ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
				defer cancel()
				inv, err := mgr.DecodePayReq(ctxt, invoiceResp.GetPaymentRequest())
				assert.NoError(t.t, err)

				nodeID, err := lnchat.NewNodeFromString(node.PubKeyStr)
				assert.NoError(t.t, err)

				expected := &lnchat.PayReq{
					Destination: nodeID,
					Amt:         lnchat.NewAmount(invoice.ValueMsat),
				}
				assert.EqualValues(t.t, expected, inv)
			},
		},
		{
			name: "Empty amount",
			test: func(net *lntest.NetworkHarness, t *harnessTest,
				mgr lnchat.LightManager, node *lntest.HarnessNode) {

				ctxb := context.Background()

				invoice := &lnrpc.Invoice{
					Memo:      "Test",
					RPreimage: bytes.Repeat([]byte("DCBA"), 8),
				}
				invoiceResp, err := node.AddInvoice(ctxb, invoice)
				if err != nil {
					t.Fatalf("unable to add invoice: %v", err)
				}

				ctxt, cancel := context.WithTimeout(ctxb, defaultTimeout)
				defer cancel()
				inv, err := mgr.DecodePayReq(ctxt, invoiceResp.GetPaymentRequest())
				assert.NoError(t.t, err)

				nodeID, err := lnchat.NewNodeFromString(node.PubKeyStr)
				assert.NoError(t.t, err)

				expected := &lnchat.PayReq{
					Destination: nodeID,
					Amt:         lnchat.NewAmount(0),
				}
				assert.EqualValues(t.t, expected, inv)
			},
		},
	}

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	for _, subTest := range subTests {
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(net, ht, mgrAlice, net.Alice)
		})

		if !success {
			break
		}
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
