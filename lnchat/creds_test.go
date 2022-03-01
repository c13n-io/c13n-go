package lnchat

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

var (
	lndHost = "127.0.0.1:10009"

	// TLS certificate body encoded as base64URL
	certBodyEnc = "" +
		"MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn-dNuaTAKBggqhkjOPQQDAjASMRAw" +
		"DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow" +
		"EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d" +
		"7VNhbWvZLWPuj_RtHFjvtJBEwOkhbN_BnnE8rnZR8-sbwnc_KhCk3FhnpHZnQz7B" +
		"5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB_wQEAwICpDATBgNVHSUEDDAKBggr" +
		"BgEFBQcDATAPBgNVHRMBAf8EBTADAQH_MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1" +
		"NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6_l" +
		"Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h_iI-i341gBmLiAFQOyTDT-_wQc" +
		"6MF9-Yw1Yy0t"

	// Macaroon bytes encoded as base64URL
	macBytesEnc = "" +
		"AgEDbG5kAvgBAwoQ0SfUsCGDIr9Q7AtZyDs3YhIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoT" +
		"CgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24S" +
		"CGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFp" +
		"bhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdy" +
		"aXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgE2S9CfguJ4T9ZIEHp4g0Ez0l2SBDNuhQ" +
		"6kEMXlRNRJ8"
)

func TestNewCredentials(t *testing.T) {
	createURL := func(host, tls, mac string) string {
		url := fmt.Sprintf("lndconnect://%s", host)

		macSeparator := "?"
		if tls != "" {
			macSeparator = "&"
			url = fmt.Sprintf("%s?cert=%s", url, tls)
		}
		if mac != "" {
			url = fmt.Sprintf("%s%smacaroon=%s", url, macSeparator, mac)
		}

		return url
	}

	mustDecodeBase64URL := func(t *testing.T, enc string) []byte {
		decoder := base64.RawURLEncoding

		bytes, err := decoder.DecodeString(enc)
		require.NoError(t, err)

		return bytes
	}

	// Recreate certificate and write to file.
	certBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: mustDecodeBase64URL(t, certBodyEnc),
	})

	tempCertPath := filepath.Join(os.TempDir(), "test_tls.cert")
	_ = ioutil.WriteFile(tempCertPath, certBytes, 0644)
	defer os.Remove(tempCertPath)

	// Write decoded macaroon to file.
	macBytes := mustDecodeBase64URL(t, macBytesEnc)

	tempMacPath := filepath.Join(os.TempDir(), "test_mac.macaroon")
	_ = ioutil.WriteFile(tempMacPath, macBytes, 0644)
	defer os.Remove(tempMacPath)

	expectedTLSCreds, err := credentials.NewClientTLSFromFile(tempCertPath, "")
	require.NoError(t, err)

	macaroon, err := loadMacaroonFromBytes(macBytes)
	require.NoError(t, err)

	expectedMacCreds := macaroonCredentials{
		Macaroon: macaroon,
	}

	expectedMacaroonVal := hex.EncodeToString(macBytes)

	cases := []struct {
		name                      string
		rpcAddr, tlsPath, macPath string
		lndConnectURL             string
		expectedCreds             lnconnect.Credentials
		expectedErr               error
	}{
		{
			name:          "lndconnect credentials without macaroon parameter",
			lndConnectURL: createURL(lndHost, certBodyEnc, ""),
			expectedCreds: lnconnect.Credentials{},
			expectedErr:   fmt.Errorf("lndconnect URL missing macaroon value"),
		},
		{
			name:          "lndconnect credentials without certificate",
			lndConnectURL: createURL(lndHost, "", macBytesEnc),
			expectedCreds: lnconnect.Credentials{},
			expectedErr:   fmt.Errorf("lndconnect URL missing cert value"),
		},
		{
			name:          "lndconnect credentials",
			lndConnectURL: createURL(lndHost, certBodyEnc, macBytesEnc),
			expectedCreds: lnconnect.Credentials{
				RPCAddress: lndHost,
				TLSCreds:   expectedTLSCreds,
				RPCCreds:   expectedMacCreds,
			},
		},
		{
			name:    "local credentials without macaroon file",
			rpcAddr: lndHost,
			tlsPath: tempCertPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress: lndHost,
				TLSCreds:   expectedTLSCreds,
			},
		},
		{
			name:    "local credentials without tls file",
			rpcAddr: lndHost,
			macPath: tempMacPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress: lndHost,
				RPCCreds:   expectedMacCreds,
			},
		},
		{
			name:    "local credentials",
			rpcAddr: lndHost,
			tlsPath: tempCertPath,
			macPath: tempMacPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress: lndHost,
				TLSCreds:   expectedTLSCreds,
				RPCCreds:   expectedMacCreds,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var creds lnconnect.Credentials
			var err error

			switch c.lndConnectURL {
			case "":
				creds, err = NewCredentials(c.rpcAddr, c.tlsPath, c.macPath)
			default:
				creds, err = NewCredentialsFromURL(c.lndConnectURL)
			}

			if c.expectedErr != nil {
				assert.EqualError(t, err, c.expectedErr.Error())
				return
			}

			assert.NoError(t, err)

			assert.EqualValues(t,
				c.expectedCreds.RPCAddress, creds.RPCAddress)
			if c.expectedCreds.TLSCreds != nil {
				assert.NotEmpty(t, creds.TLSCreds)
			}
			if c.expectedCreds.RPCCreds != nil {
				assert.NotEmpty(t, creds.RPCCreds)

				reqMetadata, err := creds.RPCCreds.GetRequestMetadata(
					context.TODO(), "",
				)
				assert.NoError(t, err)
				assert.Equal(t, reqMetadata["macaroon"], expectedMacaroonVal)
			}
		})
	}
}
