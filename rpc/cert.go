package rpc

import (
	"fmt"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/cert"
	"google.golang.org/grpc/credentials"
)

func getCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {

	// Check if the server certificate and key files exist
	if _, err := os.Stat(certFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Certificate file not found")
		}
	}
	if _, err := os.Stat(keyFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Private key file not found")
		}
	}

	// Check if the certificate has expired.
	certData, parsedCert, err := cert.LoadCert(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	if time.Now().After(parsedCert.NotAfter) {
		return nil, fmt.Errorf("Server certificate has expired")
	}

	return credentials.NewTLS(cert.TLSConfFromCert(certData)), nil
}
