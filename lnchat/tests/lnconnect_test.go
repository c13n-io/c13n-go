package itest

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	netutils "net"
	"strconv"
	"testing"

	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"

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
			name:       "no macaroon credentials provided",
			testType:   "short",
			tlsPath:    net.Alice.TLSCertStr(),
			macPath:    "",
			rpcAddress: net.Alice.Cfg.RPCAddr(),
			err:        nil,
		},
	}

	for _, c := range cases {
		t.t.Run(c.name, func(t *testing.T) {
			if c.testType == "long" && testing.Short() {
				t.Skip("skipping test in short mode.")
			}

			var tls credentials.TransportCredentials
			var mac credentials.PerRPCCredentials
			var err error

			if c.tlsPath != "" {
				tls, err = loadTLSCert(c.tlsPath)
				assert.NoError(t, err)
			}
			if c.macPath != "" {
				mac, err = loadMacaroonCreds(c.macPath)
				assert.NoError(t, err)
			}

			cfg := lnconnect.Credentials{
				RPCAddress: c.rpcAddress,
				TLSCreds:   tls,
				RPCCreds:   mac,
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

func loadMacaroonCreds(macPath string) (credentials.PerRPCCredentials, error) {
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		return nil, err
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		return nil, err
	}

	return testMacaroonCredentials{
		Macaroon: mac,
	}, nil
}

type testMacaroonCredentials struct {
	*macaroon.Macaroon
}

func (c testMacaroonCredentials) RequireTransportSecurity() bool {
	return true
}

func (c testMacaroonCredentials) GetRequestMetadata(_ context.Context,
	_ ...string) (map[string]string, error) {

	macBytes, err := c.MarshalBinary()
	if err != nil {
		return nil, err
	}

	md := make(map[string]string)
	md["macaroon"] = hex.EncodeToString(macBytes)

	return md, nil
}
