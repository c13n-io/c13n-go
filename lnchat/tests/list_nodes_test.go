package itest

import (
	"context"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/stretchr/testify/assert"
)

func testListNodes(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	singleHopSubTests := []testCase{
		{
			name: "Two nodes, no channels",
			test: testListNodesNoChannel,
		},
		{
			name: "Two nodes, one channel (A -> B)",
			test: testListNodesOneChannel,
		},
		{
			name: "Three nodes, two channels (A -> B -> C)",
			test: testListNodesTwoChannels,
		},
	}

	for _, subTest := range singleHopSubTests {
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

// Test ListNodes on a network with nodes Alice and Bob and no channels.
// Topology: (A, B)
func testListNodesNoChannel(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	// List visible nodes (should be only one's own node).
	list, err := mgrAlice.ListNodes(context.Background())
	assert.NoError(t.t, err)
	assert.Lenf(t.t, list, 1, "ListNodes reported #nodes = %d (!= 1) for A on topology A, B", len(list))

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

// Test ListNodes on a network with nodes Alice and Bob and
// one channel from Alice to Bob.
// Topology: (A -> B)
func testListNodesOneChannel(net *lntest.NetworkHarness, t *harnessTest) {
	ctxb := context.Background()

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	// Open channel with Bob.
	chanParams := lntest.OpenChannelParams{
		Amt:     1000000,
		Private: false,
	}
	chanPoint := openChannelAndAssert(t, net, net.Alice, net.Bob, chanParams)

	err = net.Alice.WaitForNetworkChannelOpen(ctxb, chanPoint)
	assert.NoErrorf(t.t, err, "alice didn't report channel")

	err = net.Bob.WaitForNetworkChannelOpen(ctxb, chanPoint)
	assert.NoErrorf(t.t, err, "bob didn't report channel")

	// List visible nodes (should be 2: Alice and Bob).
	list, err := mgrAlice.ListNodes(ctxb)
	assert.NoError(t.t, err)
	assert.Lenf(t.t, list, 2, "ListNodes reported #nodes = %d (!= 2) for A on topology A->B", len(list))

	err = mgrAlice.Close()
	assert.NoError(t.t, err)

	if err := wait.NoError(
		assertNumPendingHTLCs(0, net.Alice, net.Bob),
		pendingHTLCTimeout,
	); err != nil {
		t.Fatalf("Unable to assert no pending htlcs: %v", err)
	}

	// Close Alice -> Bob channel.
	closeChannelAndAssert(t, net, net.Alice, chanPoint, false)
}

// Test ListNodes on a network with nodes Alice and Bob,
// one channel from Alice to Bob and one from Bob to Carol.
// Topology: (A -> B -> C)
func testListNodesTwoChannels(net *lntest.NetworkHarness, t *harnessTest) {
	ctxb := context.Background()

	// The network topology should be A -> B -> C after this call.
	aliceBobChanPoint, bobCarolChanPoint, carol := createThreeHopNetwork(
		t, net, net.Alice, net.Bob, false,
		lnrpc.CommitmentType_STATIC_REMOTE_KEY,
	)

	mgrCarol, err := createNodeManager(carol)
	assert.NoError(t.t, err)

	// List visible nodes (should be 3: Alice, Bob and Carol).
	list, err := mgrCarol.ListNodes(ctxb)
	assert.NoError(t.t, err)
	assert.Lenf(t.t, list, 3, "ListNodes reported #nodes = %d (!= 3) for A on topology A->B->C", len(list))

	err = mgrCarol.Close()
	assert.NoError(t.t, err)

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
}
