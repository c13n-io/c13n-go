package itest

import (
	"encoding/base64"
	"fmt"

	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-backend/lnchat"
)

func testNewLnchat(net *lntest.NetworkHarness, t *harnessTest) {

	manager, err := lnchat.New(net.Alice.Cfg.RPCAddr(),
		lnchat.WithMacaroonPath(net.Alice.AdminMacPath()),
		lnchat.WithTLSPath(net.Alice.TLSCertStr()))
	assert.NoError(t.t, err)

	err = manager.Close()
	assert.NoError(t.t, err)
}

func testNewLnchatFromLNDConnectURL(net *lntest.NetworkHarness, t *harnessTest) {
	// First construct the lndconnecturl
	tlsBytes, err := readTLSBytes(net.Alice.TLSCertStr())
	assert.NoError(t.t, err)

	macBytes, err := readMacaroonBytes(net.Alice.AdminMacPath())
	assert.NoError(t.t, err)

	tlsStringEnc := base64.RawURLEncoding.EncodeToString(tlsBytes)
	macStringEnc := base64.RawURLEncoding.EncodeToString(macBytes)

	lndconnecturl := fmt.Sprintf("lndconnect://%s?cert=%s&macaroon=%s",
		net.Alice.Cfg.RPCAddr(), tlsStringEnc, macStringEnc)
	manager, err := lnchat.NewFromURL(lndconnecturl)
	assert.NoError(t.t, err)

	err = manager.Close()
	assert.NoError(t.t, err)
}
