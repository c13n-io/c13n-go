package itest

import (
	"github.com/lightningnetwork/lnd/lntest"

	"github.com/c13n-io/c13n-go/lnchat"
)

func createNodeManager(node *lntest.HarnessNode) (lnchat.LightManager, error) {
	return lnchat.New(node.Cfg.RPCAddr(),
		lnchat.WithMacaroonPath(node.AdminMacPath()),
		lnchat.WithTLSPath(node.TLSCertStr()))
}
