package itest

import (
	"github.com/lightningnetwork/lnd/lntest"

	"github.com/c13n-io/c13n-go/lnchat"
)

func createNodeManager(node *lntest.HarnessNode) (lnchat.LightManager, error) {
	creds, err := lnchat.NewCredentials(
		node.Cfg.RPCAddr(),
		node.TLSCertStr(),
		node.AdminMacPath(),
	)
	if err != nil {
		return nil, err
	}

	return lnchat.New(creds)
}
