package lnconnect

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ErrCredentials represents a credentials error.
var ErrCredentials = fmt.Errorf("credentials error")

// Credentials is used to pass identification and connection
// parameters to InitializeConnection.
type Credentials struct {
	RPCAddress string
	TLSCreds   credentials.TransportCredentials
	RPCCreds   credentials.PerRPCCredentials
}

var (
	dialTimeoutSeconds = 10

	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 50)
)

// InitializeConnection establishes a connection with a Lightning daemon.
func InitializeConnection(cfg Credentials) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.TLSCreds),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(maxMsgRecvSize),
	}

	if cfg.RPCCreds != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(cfg.RPCCreds))
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
