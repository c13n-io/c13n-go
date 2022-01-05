package lnchat

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

var (
	certContents = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`

	certBody = "" +
		"MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn-dNuaTAKBggqhkjOPQQDAjASMRAw" +
		"DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow" +
		"EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d" +
		"7VNhbWvZLWPuj_RtHFjvtJBEwOkhbN_BnnE8rnZR8-sbwnc_KhCk3FhnpHZnQz7B" +
		"5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB_wQEAwICpDATBgNVHSUEDDAKBggr" +
		"BgEFBQcDATAPBgNVHRMBAf8EBTADAQH_MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1" +
		"NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6_l" +
		"Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h_iI-i341gBmLiAFQOyTDT-_wQc" +
		"6MF9-Yw1Yy0t"

	macEncBase64 = "" +
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

	mustDecodeBase64 := func(t *testing.T, enc string) []byte {
		decoder := base64.RawURLEncoding

		bytes, err := decoder.DecodeString(enc)
		require.NoError(t, err)

		return bytes
	}

	lndHost := "127.0.0.1:10009"

	tlsBytes := mustDecodeBase64(t, certBody)
	macBytes := mustDecodeBase64(t, macEncBase64)

	tempCertPath := filepath.Join(os.TempDir(), "test_tls.cert")
	_ = ioutil.WriteFile(tempCertPath, []byte(certContents), 0644)
	defer os.Remove(tempCertPath)

	tempMacPath := filepath.Join(os.TempDir(), "test_mac.macaroon")
	_ = ioutil.WriteFile(tempMacPath, macBytes, 0644)
	defer os.Remove(tempMacPath)

	cases := []struct {
		name                      string
		rpcAddr, tlsPath, macPath string
		lndConnectURL             string
		expectedCreds             lnconnect.Credentials
		expectedErr               error
	}{
		{
			name:          "lndconnect credentials without macaroon parameter",
			lndConnectURL: createURL(lndHost, certBody, ""),
			expectedCreds: lnconnect.Credentials{},
			expectedErr:   fmt.Errorf("macaroon must be present in lndconnect URL"),
		},
		{
			name:          "lndconnect credentials without cert parameter",
			lndConnectURL: createURL(lndHost, "", macEncBase64),
			expectedCreds: lnconnect.Credentials{},
			expectedErr:   fmt.Errorf("TLS certificate must be present in lndconnect URL"),
		},
		{
			name:          "lndconnect credentials",
			lndConnectURL: createURL(lndHost, certBody, macEncBase64),
			expectedCreds: lnconnect.Credentials{
				RPCAddress:    lndHost,
				TLSBytes:      tlsBytes,
				MacaroonBytes: macBytes,
			},
		},
		{
			name:    "local credentials without macaroon file",
			rpcAddr: lndHost,
			tlsPath: tempCertPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress: lndHost,
				TLSBytes:   tlsBytes,
			},
		},
		{
			name:    "local credentials without tls file",
			rpcAddr: lndHost,
			macPath: tempMacPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress:    lndHost,
				MacaroonBytes: macBytes,
			},
		},
		{
			name:    "local credentials",
			rpcAddr: lndHost,
			tlsPath: tempCertPath,
			macPath: tempMacPath,
			expectedCreds: lnconnect.Credentials{
				RPCAddress:    lndHost,
				TLSBytes:      tlsBytes,
				MacaroonBytes: macBytes,
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

			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.NotEmpty(t, creds)
				assert.EqualValues(t, c.expectedCreds, creds)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}
