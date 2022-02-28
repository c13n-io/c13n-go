package lnchat

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

// NewCredentials constructs a set of credentials
// from the address and the TLS and macaroon file paths.
func NewCredentials(rpcAddr, tlsPath, macPath string) (lnconnect.Credentials, error) {
	creds := lnconnect.Credentials{
		RPCAddress: rpcAddr,
	}

	var err error
	if tlsPath != "" {
		creds.TLSCreds, err = loadTLSCreds(tlsPath)
		if err != nil {
			return creds, errors.Wrap(err, "could not read TLS file")
		}
	}

	if macPath != "" {
		creds.MacaroonBytes, err = loadMacFile(macPath)
		if err != nil {
			return creds, errors.Wrap(err, "could not read macaroon file")
		}
	}

	return creds, nil
}

// NewCredentialsFromURL constructs a set of credentials
// from an lndconnect URL.
func NewCredentialsFromURL(lndConnectURL string) (lnconnect.Credentials, error) {
	return parseLNDConnectURL(lndConnectURL)
}

func parseLNDConnectURL(lndConnectURL string) (lnconnect.Credentials, error) {
	var creds lnconnect.Credentials

	if lndConnectURL != "" {
		lndURL, err := url.Parse(lndConnectURL)
		if err != nil {
			return creds, errors.Wrap(err, "could not parse lndconnect URL")
		}
		if lndURL.Scheme != "lndconnect" {
			return creds, errors.New("invalid scheme for lndconnect URL")
		}

		creds.RPCAddress = lndURL.Host

		queryMap := lndURL.Query()
		decoder := base64.RawURLEncoding

		// If the query map does not contain a cert key or it's empty, error out.
		cert, ok := queryMap["cert"]
		if !ok || len(cert) != 1 || cert[0] == "" {
			return creds, errors.New("TLS certificate must be present in lndconnect URL")
		}
		tlsBytes, err := decoder.DecodeString(cert[0])
		if err != nil {
			return creds, errors.Wrap(err, "could not decode TLS bytes")
		}
		creds.TLSCreds, err = loadTLSCredsFromBytes(
			pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: tlsBytes,
			}),
		)
		if err != nil {
			return creds, err
		}

		mac, ok := queryMap["macaroon"]
		if !ok || len(mac) != 1 || mac[0] == "" {
			return creds, errors.New("macaroon must be present in lndconnect URL")
		}
		creds.MacaroonBytes, err = decoder.DecodeString(mac[0])
		if err != nil {
			return creds, errors.Wrap(err, "could not decode macaroon bytes")
		}
	}

	return creds, nil
}

func loadTLSCreds(tlsPath string) (credentials.TransportCredentials, error) {
	certBytes, err := ioutil.ReadFile(tlsPath)
	if err != nil {
		return nil, err
	}

	return loadTLSCredsFromBytes(certBytes)
}

func loadTLSCredsFromBytes(certBytes []byte) (credentials.TransportCredentials, error) {
	tlsCertPool := x509.NewCertPool()
	if !tlsCertPool.AppendCertsFromPEM(certBytes) {
		return nil, errors.New("failed to append certificate")
	}

	return credentials.NewClientTLSFromCert(tlsCertPool, ""), nil
}

func loadMacFile(macaroonPath string) ([]byte, error) {
	macBytes, err := ioutil.ReadFile(macaroonPath)

	return macBytes, err
}
