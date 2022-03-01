package lnchat

import (
	"context"
	"encoding/hex"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
)

// macaroonCredentials implements grpc/credentials.PerRPCCredentials interface.
type macaroonCredentials struct {
	*macaroon.Macaroon
}

func (mc macaroonCredentials) RequireTransportSecurity() bool {
	return true
}

func (mc macaroonCredentials) GetRequestMetadata(_ context.Context,
	_ ...string) (map[string]string, error) {

	macBytes, err := mc.Macaroon.MarshalBinary()
	if err != nil {
		return nil, err
	}

	md := make(map[string]string)
	md["macaroon"] = hex.EncodeToString(macBytes)

	return md, nil
}

// Compile-time assertion to ensure macaroonCredentials
// implements credentials.PerRPCCredentials interface.
var _ credentials.PerRPCCredentials = macaroonCredentials{}

func loadMacaroon(macPath string) (*macaroon.Macaroon, error) {
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		return nil, err
	}

	return loadMacaroonFromBytes(macBytes)
}

func loadMacaroonFromBytes(macBytes []byte) (*macaroon.Macaroon, error) {
	mac := &macaroon.Macaroon{}

	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return nil, err
	}

	return mac, nil
}
