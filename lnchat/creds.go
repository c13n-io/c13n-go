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
func NewCredentials(rpcAddr, tlsPath, macPath string,
	macConstraints MacaroonConstraints) (lnconnect.Credentials, error) {

	creds := lnconnect.Credentials{
		RPCAddress: rpcAddr,
	}

	if tlsPath != "" {
		tlsCreds, err := loadTLSCreds(tlsPath)
		if err != nil {
			return creds, errors.Wrap(err, "could not load TLS certificate")
		}

		creds.TLSCreds = tlsCreds
	}

	if macPath != "" {
		mac, err := loadMacaroon(macPath)
		if err != nil {
			return creds, errors.Wrap(err, "could not load macaroon")
		}

		creds.RPCCreds = macaroonCredentials{
			Macaroon:    mac,
			constraints: macConstraints,
		}
	}

	return creds, nil
}

// NewCredentialsFromURL constructs a set of credentials
// from an lndconnect URL.
func NewCredentialsFromURL(lndConnectURL string,
	macConstraints MacaroonConstraints) (lnconnect.Credentials, error) {
	creds := lnconnect.Credentials{}

	addr, tlsBytes, macBytes, err := parseLNDConnectURL(lndConnectURL)
	if err != nil {
		return creds, err
	}
	creds.RPCAddress = addr

	tlsCreds, err := loadTLSCredsFromBytes(
		pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: tlsBytes,
		}),
	)
	if err != nil {
		return creds, errors.Wrap(err, "could not load TLS certificate")
	}
	creds.TLSCreds = tlsCreds

	mac, err := loadMacaroonFromBytes(macBytes)
	if err != nil {
		return creds, errors.Wrap(err, "could not load macaroon")
	}
	creds.RPCCreds = macaroonCredentials{
		Macaroon:    mac,
		constraints: macConstraints,
	}

	return creds, nil
}

func parseLNDConnectURL(lndConnectURL string) (string, []byte, []byte, error) {
	decoder := base64.RawURLEncoding
	// Retrieve and decode an attribute (base64URL-encoded) from a url query.
	getDecodedAttrValue := func(values url.Values, key string) ([]byte, error) {
		val := values.Get(key)
		decodedVal, err := decoder.DecodeString(val)
		if err != nil {
			return nil, err
		}

		return decodedVal, nil
	}

	var addr string
	var tlsBytes, macBytes []byte
	var err error
	if lndConnectURL != "" {
		lndURL, err := url.Parse(lndConnectURL)
		if err != nil {
			return "", nil, nil,
				errors.Wrap(err, "could not parse lndconnect URL")
		}
		if lndURL.Scheme != "lndconnect" {
			return "", nil, nil,
				errors.New("invalid scheme for lndconnect URL")
		}

		addr = lndURL.Host
		attrMap := lndURL.Query()

		tlsBytes, err = getDecodedAttrValue(attrMap, "cert")
		if err != nil {
			return addr, nil, nil,
				errors.Wrap(err, "could not decode certificate")
		}
		if len(tlsBytes) == 0 {
			return addr, nil, nil,
				errors.New("lndconnect URL missing cert value")
		}

		macBytes, err = getDecodedAttrValue(attrMap, "macaroon")
		if err != nil {
			return addr, tlsBytes, nil,
				errors.Wrap(err, "could not decode macaroon")
		}
		if len(macBytes) == 0 {
			return addr, tlsBytes, nil,
				errors.New("lndconnect URL missing macaroon value")
		}
	}

	return addr, tlsBytes, macBytes, err
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
