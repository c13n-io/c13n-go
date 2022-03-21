package itest

import (
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testNewLnchat(net *lntest.NetworkHarness, t *harnessTest) {
	constraints := lnchat.MacaroonConstraints{}
	creds, err := lnchat.NewCredentials(
		net.Alice.Cfg.RPCAddr(),
		net.Alice.TLSCertStr(),
		net.Alice.AdminMacPath(),
		constraints,
	)
	require.NoError(t.t, err)

	manager, err := lnchat.New(creds)
	assert.NoError(t.t, err)

	err = manager.Close()
	assert.NoError(t.t, err)
}
