package lnchat

import (
	"encoding/base64"
	"net/url"

	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

func parseLNDConnectURL(lndConnectURL string) (*lnconnect.Credentials, error) {
	var tlsBytes, macaroonBytes []byte
	var hostPort string

	if lndConnectURL != "" {
		lndURL, err := url.Parse(lndConnectURL)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse lndconnect URL")
		}
		if lndURL.Scheme != "lndconnect" {
			return nil, errors.New("invalid scheme for lndconnect URL")
		}

		hostPort = lndURL.Host

		queryMap := lndURL.Query()
		decoder := base64.RawURLEncoding

		// If the query map does not contain a cert key or it's empty, error out.
		cert, ok := queryMap["cert"]
		if !ok || len(cert) != 1 || cert[0] == "" {
			return nil, errors.New("TLS certificate must be present in lndconnect URL")
		}
		if tlsBytes, err = decoder.DecodeString(cert[0]); err != nil {
			return nil, errors.Wrap(err, "could not decode TLS bytes")
		}

		mac, ok := queryMap["macaroon"]
		if !ok || len(mac) != 1 || mac[0] == "" {
			return nil, errors.New("macaroon must be present in lndconnect URL")
		}
		if macaroonBytes, err = decoder.DecodeString(mac[0]); err != nil {
			return nil, errors.Wrap(err, "could not decode macaroon bytes")
		}
	}

	return &lnconnect.Credentials{
		TLSBytes:      tlsBytes,
		MacaroonBytes: macaroonBytes,
		RPCAddress:    hostPort,
	}, nil
}
