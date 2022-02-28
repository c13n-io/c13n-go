package itest

import (
	"context"
	"fmt"
	"io/ioutil"
	netutils "net"
	"strconv"
	"testing"

	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
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
			name:       "success",
			testType:   "short",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    net.Alice.AdminMacPath(),
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err:        nil,
		},
		{
			name:     "connection timeout",
			testType: "long",
			tlsPath:  net.Alice.TLSCertStr(),
			macPath:  net.Alice.AdminMacPath(),
			rpcAddress: netutils.JoinHostPort("127.0.0.1",
				strconv.Itoa(net.Alice.Cfg.RPCPort-42)),
			err: context.DeadlineExceeded,
		},
		{
			name:       "invalid transport credentials",
			testType:   "short",
			tlsPath:    "",
			macPath:    net.Alice.AdminMacPath(),
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err: fmt.Errorf("credentials error: " +
				"TLS certificate not provided"),
		},
		{
			name:       "invalid macaroon credentials",
			testType:   "short",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    "",
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err: fmt.Errorf("credentials error: " +
				"could not unmarshal macaroon: empty macaroon data"),
		},
	}

	for _, c := range cases {
		t.t.Run(c.name, func(t *testing.T) {
			if c.testType == "long" && testing.Short() {
				t.Skip("skipping test in short mode.")
			}

			var tls credentials.TransportCredentials
			var mac []byte
			var err error

			if c.tlsPath != "" {
				tls, err = loadTLSCert(c.tlsPath)
				assert.NoError(t, err)
			}
			if c.macPath != "" {
				mac, err = readMacaroonBytes(c.macPath)
				assert.NoError(t, err)
			} else {
				mac = []byte{}
			}

			cfg := lnconnect.Credentials{
				MacaroonBytes: mac,
				TLSCreds:      tls,
				RPCAddress:    c.rpcAddress,
			}

			_, err = lnconnect.InitializeConnection(cfg)

			switch c.err {
			case nil:
				assert.NoError(t, err)
			default:
				assert.EqualError(t, err, c.err.Error())
			}
		})
	}
}

func loadTLSCert(tlsPath string) (credentials.TransportCredentials, error) {
	return credentials.NewClientTLSFromFile(tlsPath, "")
}

func readMacaroonBytes(macPath string) ([]byte, error) {
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		return nil, err
	}

	return macBytes, nil
}
