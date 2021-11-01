package lnchat

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/macaroon.v2"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

func TestParseLNDConnectURL(t *testing.T) {
	rpcAddr := "127.0.0.1:1111"

	mac, err := macaroon.New([]byte("root_key"), []byte("0"),
		"test", macaroon.LatestVersion)
	assert.NoError(t, err)
	macBytes, err := mac.MarshalBinary()
	assert.NoError(t, err)

	certPem := `-----BEGIN CERTIFICATE-----
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
	block, _ := pem.Decode([]byte(certPem))
	tlsBytes := block.Bytes

	tempTLSPath := filepath.Join(".", "tls_test.cert")
	tempMacPath := filepath.Join(".", "mac_test.macaroon")
	_ = ioutil.WriteFile(tempTLSPath, []byte(certPem), 0644)
	_ = ioutil.WriteFile(tempMacPath, macBytes, 0644)
	defer func() {
		os.Remove(tempTLSPath)
		os.Remove(tempMacPath)
	}()

	macStringEnc := base64.RawURLEncoding.EncodeToString(macBytes)
	tlsStringEnc := base64.RawURLEncoding.EncodeToString(tlsBytes)

	cases := []struct {
		name           string
		lndconnecturl  string
		expectedOutput *lnconnect.Credentials
		expectedError  error
	}{
		{
			name: "TLS and Macaroon from lndconnect URL",
			lndconnecturl: fmt.Sprintf("lndconnect://%s?cert=%s&macaroon=%s",
				rpcAddr, tlsStringEnc, macStringEnc),
			expectedOutput: &lnconnect.Credentials{
				TLSBytes:      tlsBytes,
				MacaroonBytes: macBytes,
				RPCAddress:    rpcAddr,
			},
			expectedError: nil,
		},
		{
			name: "TLS missing from lndconnect URL",
			lndconnecturl: fmt.Sprintf("lndconnect://%s?macaroon=%s",
				rpcAddr, macStringEnc),
			expectedOutput: nil,
			expectedError:  errors.New("TLS certificate must be present in lndconnect URL"),
		},
		{
			name: "Macaroon missing from lndconnect URL",
			lndconnecturl: fmt.Sprintf("lndconnect://%s?cert=%s",
				rpcAddr, tlsStringEnc),
			expectedOutput: nil,
			expectedError:  errors.New("macaroon must be present in lndconnect URL"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			config, err := parseLNDConnectURL(c.lndconnecturl)

			if c.expectedError != nil {
				assert.EqualError(t, err, c.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.EqualValues(t, config, c.expectedOutput)
			}
		})
	}
}
