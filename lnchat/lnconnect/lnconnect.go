package lnconnect

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
)

// ErrCredentials represents a credentials error.
var ErrCredentials = fmt.Errorf("Credentials error")

// Credentials is used to pass identification and connection
// parameters to InitializeConnection.
type Credentials struct {
	TLSBytes      []byte
	MacaroonBytes []byte
	RPCAddress    string
}

var (
	dialTimeoutSeconds = 10

	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 50)
)

// InitializeConnection establishes a connection with a Lightning daemon.
func InitializeConnection(cfg Credentials) (*grpc.ClientConn, error) {
	certBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cfg.TLSBytes,
	})
	if certBytes == nil {
		return nil, fmt.Errorf("%w: could not encode TLS certificate", ErrCredentials)
	}

	tlsCerts := x509.NewCertPool()
	if !tlsCerts.AppendCertsFromPEM(certBytes) {
		return nil, fmt.Errorf("%w: could not append TLS certificate to pool", ErrCredentials)
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(cfg.MacaroonBytes); err != nil {
		return nil, fmt.Errorf("%w: could not unmarshal macaroon: %v", ErrCredentials, err)
	}

	// Use a constrained macaroon for RPC calls
	macConstraints := []macaroons.Constraint{
		// Lock macaroon to our IP address (empty string means no caveat)
		macaroons.IPLockConstraint(""),
		// Define macaroon timeout in seconds (max-int64 overflows?)
		macaroons.TimeoutConstraint(1 << 32),
	}
	constrMac, err := macaroons.AddConstraints(mac, macConstraints...)
	if err != nil {
		return nil, fmt.Errorf("%w: could not add macaroon constraints: %v",
			ErrCredentials, err)
	}
	perRPCCreds, err := macaroons.NewMacaroonCredential(constrMac)
	if err != nil {
		return nil, fmt.Errorf("%w: could not create per-RPC credentials: %v",
			ErrCredentials, err)
	}

	config := &tls.Config{}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithPerRPCCredentials(perRPCCreds),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(maxMsgRecvSize),
	}

	// Set a timeout for the blocking connection call
	dialCtx, cancel := context.WithTimeout(context.Background(),
		time.Duration(dialTimeoutSeconds)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, cfg.RPCAddress, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
