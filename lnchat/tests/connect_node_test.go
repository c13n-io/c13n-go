package itest

import (
	"context"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/stretchr/testify/assert"
)

const DefaultTimeout = time.Duration(30) * time.Second

func testConnectNode(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "Not connected",
			test: testConnectNodeNotConnected,
		},
		{
			name: "Already connected",
			test: testConnectNodeAlreadyConnected,
		},
	}

	for _, subTest := range subTests {
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

func findTargetInPeerList(src, target *lntest.HarnessNode) bool {
	// If node B is seen in the ListPeers response from node A,
	// then we can exit early as the connection has been fully
	// established.
	ctxt, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	resp, err := target.ListPeers(ctxt, &lnrpc.ListPeersRequest{})
	if err != nil {
		return false
	}

	for _, peer := range resp.Peers {
		if peer.PubKey == src.PubKeyStr {
			return true
		}
	}

	return false
}

func testConnectNodeNotConnected(net *lntest.NetworkHarness, t *harnessTest) {

	err := t.lndHarness.DisconnectNodes(net.Alice, net.Bob)
	assert.NoError(t.t, err)

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	err = mgrAlice.ConnectNode(context.Background(), net.Bob.PubKeyStr, net.Bob.Cfg.P2PAddr())
	assert.NoError(t.t, err)

	err = wait.Predicate(func() bool {
		return findTargetInPeerList(net.Alice, net.Bob) &&
			findTargetInPeerList(net.Bob, net.Alice)
	}, DefaultTimeout)
	assert.NoError(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}

func testConnectNodeAlreadyConnected(net *lntest.NetworkHarness, t *harnessTest) {

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	err = mgrAlice.ConnectNode(context.Background(), net.Bob.PubKeyStr, net.Bob.Cfg.P2PAddr())
	assert.NoError(t.t, err)

	err = wait.Predicate(func() bool {
		return findTargetInPeerList(net.Alice, net.Bob) &&
			findTargetInPeerList(net.Bob, net.Alice)
	}, DefaultTimeout)
	assert.NoError(t.t, err)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
