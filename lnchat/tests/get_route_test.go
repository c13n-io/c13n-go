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
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, payOpts, payload)

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
			name: "Intermmediate, with fees",
			test: testGetRouteMultiHopWithFees,
		},
		{
			name: "No route found",
			test: testGetRouteMultiHopNoRouteFound,
		},
	}

	aliceBobChanPoint, bobCarolChanPoint, carol := createThreeHopNetwork(t,
		net, net.Alice, net.Bob, false,
		lnrpc.CommitmentType_STATIC_REMOTE_KEY,
	)

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
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, payOpts, payload)

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
	response, _, err := mgrAlice.GetRoute(ctxt, recipient, amount, payOpts, payload)

	assert.Nil(t.t, response)
	assert.Error(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
