package itest

import (
	"encoding/pem"
	"errors"
	"io/ioutil"
	netutils "net"
	"strconv"
	"testing"

	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-backend/lnchat/lnconnect"
)

func testInitializeConnection(net *lntest.NetworkHarness, t *harnessTest) {
	cases := []struct {
		name       string
		testType   string
		tlsPath    string
		macPath    string
		rpcAddress string
		err        error
	}{
		{
			name:       "Success",
			testType:   "short",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    net.Alice.AdminMacPath(),
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err:        nil,
		},
		{
			name:       "TimeOut",
			testType:   "long",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    net.Alice.AdminMacPath(),
			rpcAddress: netutils.JoinHostPort("127.0.0.1", strconv.Itoa(net.Alice.Cfg.RPCPort-42)),
			err:        assert.AnError,
		},
		{
			name:       "Invalid TLS",
			testType:   "short",
			tlsPath:    "",
			macPath:    net.Alice.AdminMacPath(),
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err:        lnconnect.ErrCredentials,
		},
		{
			name:       "Invalid Macaroon path",
			testType:   "short",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    "",
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err:        lnconnect.ErrCredentials,
		},
	}

	for _, c := range cases {
		t.t.Run(c.name, func(t *testing.T) {
			if c.testType == "long" && testing.Short() {
				t.Skip("skipping test in short mode.")
			}
			var tls []byte
			var mac []byte
			var err error

			if c.tlsPath != "" {
				tls, err = readTLSBytes(c.tlsPath)
				assert.NoError(t, err)
			} else {
				tls = []byte{}
			}
			if c.macPath != "" {
				mac, err = readMacaroonBytes(c.macPath)
				assert.NoError(t, err)
			} else {
				mac = []byte{}
			}

			cfg := lnconnect.Credentials{
				TLSBytes:      tls,
				MacaroonBytes: mac,
				RPCAddress:    c.rpcAddress,
			}

			_, err = lnconnect.InitializeConnection(cfg)

			if c.err == nil {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) && c.err != assert.AnError {
					assert.True(t, errors.Is(err, c.err))
				}
			}
		})
	}
}

func readTLSBytes(tlsPath string) ([]byte, error) {
	tlsBytes, err := ioutil.ReadFile(tlsPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(tlsBytes)

	return block.Bytes, nil
}

func readMacaroonBytes(macPath string) ([]byte, error) {
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		return nil, err
	}

	return macBytes, nil
}
