package lnchat

import (
	"context"
	"encoding/hex"
	"io/ioutil"

	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
)

// MacaroonConstraints represents constraints
// used to derive the macaroon sent for requests.
type MacaroonConstraints struct {
	// Timeout restricts the lifetime of
	// the transmitted macaroon (in seconds).
	// A value of 0 (default) does not set a timeout caveat.
	Timeout int64
	// IPLock locks the transmitted macaroon to an IP address.
	// Empty value is ignored.
	IPLock string
}

// macaroonCredentials implements grpc/credentials.PerRPCCredentials interface.
type macaroonCredentials struct {
	*macaroon.Macaroon
	constraints MacaroonConstraints
}

func (mc macaroonCredentials) RequireTransportSecurity() bool {
	return true
}

func (mc macaroonCredentials) GetRequestMetadata(_ context.Context,
	_ ...string) (map[string]string, error) {

	constrained, err := mc.deriveConstrained()
	if err != nil {
		return nil, err
	}

	macBytes, err := constrained.MarshalBinary()
	if err != nil {
		return nil, err
	}

	md := make(map[string]string)
	md["macaroon"] = hex.EncodeToString(macBytes)

	return md, nil
}

func (mc macaroonCredentials) deriveConstrained() (*macaroon.Macaroon, error) {
	constraints := []macaroons.Constraint{
		macaroons.IPLockConstraint(mc.constraints.IPLock),
	}

	if mc.constraints.Timeout != 0 {
		constraints = append(constraints,
			macaroons.TimeoutConstraint(mc.constraints.Timeout))
	}

	restrictedMac, err := macaroons.AddConstraints(mc.Macaroon, constraints...)
	if err != nil {
		return nil, errors.Wrap(err, "could not add macaroon constraints")
	}

	return restrictedMac, nil
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
