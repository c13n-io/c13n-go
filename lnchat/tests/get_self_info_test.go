package itest

import (
	"context"

	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-backend/lnchat"
)

func testGetSelfInfo(net *lntest.NetworkHarness, t *harnessTest) {
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	infoAlice, err := mgrAlice.GetSelfInfo(context.Background())
	assert.NoError(t.t, err)
	assert.Lenf(t.t, infoAlice.Node.Address, 2*33, "GetSelfInfo reported "+
		"len(Address) = %d (!= 66)", len(infoAlice.Node.Address))

	expectedChains := []lnchat.Chain{
		{
			Chain:   "bitcoin",
			Network: "regtest",
		},
	}
	assert.EqualValues(t.t, expectedChains, infoAlice.Chains)

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
