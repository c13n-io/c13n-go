package lnchat

import (
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/macaroon.v2"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

func TestWithTLSPath(t *testing.T) {
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

	_ = ioutil.WriteFile(tempTLSPath, []byte(certPem), 0644)

	defer func() {
		os.Remove(tempTLSPath)
	}()

	cases := []struct {
		name            string
		tlsPath         string
		expectedManager manager
		expectedError   error
	}{
		{
			name:    "TLS path success",
			tlsPath: tempTLSPath,
			expectedManager: manager{
				creds: lnconnect.Credentials{
					TLSBytes: tlsBytes,
				},
			},
			expectedError: nil,
		},
		{
			name:            "TLS path empty",
			tlsPath:         "",
			expectedManager: manager{},
			expectedError:   errors.New("TLS path empty"),
		},
		{
			name:            "TLS path not existing",
			tlsPath:         "non_existing_path",
			expectedManager: manager{},
			expectedError:   errors.New("could not read TLS file: open non_existing_path: no such file or directory"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			optionFunc := WithTLSPath(c.tlsPath)
			mgr := &manager{
				creds: lnconnect.Credentials{},
			}

			err := optionFunc(mgr)

			if c.expectedError != nil {
				assert.EqualError(t, err, c.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mgr.creds.TLSBytes)
				assert.EqualValues(t, c.expectedManager.creds.TLSBytes, mgr.creds.TLSBytes)
			}
		})
	}
}

func TestWithMacaroonPath(t *testing.T) {
	mac, err := macaroon.New([]byte("root_key"), []byte("0"),
		"test", macaroon.LatestVersion)
	assert.NoError(t, err)
	macBytes, err := mac.MarshalBinary()
	assert.NoError(t, err)

	tempMacPath := filepath.Join(".", "mac_test.macaroon")
	_ = ioutil.WriteFile(tempMacPath, macBytes, 0644)
	defer func() {
		os.Remove(tempMacPath)
	}()

	cases := []struct {
		name            string
		macaroonPath    string
		expectedManager manager
		expectedError   error
	}{
		{
			name:         "Macaroon path success",
			macaroonPath: tempMacPath,
			expectedManager: manager{
				creds: lnconnect.Credentials{
					MacaroonBytes: macBytes,
				},
			},
			expectedError: nil,
		},
		{
			name:            "Macaroon path empty",
			macaroonPath:    "",
			expectedManager: manager{},
			expectedError:   errors.New("Macaroon path empty"),
		},
		{
			name:            "Macaroon path not existing",
			macaroonPath:    "non_existing_path",
			expectedManager: manager{},
			expectedError:   errors.New("could not read macaroon file: open non_existing_path: no such file or directory"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			optionFunc := WithMacaroonPath(c.macaroonPath)
			mgr := &manager{
				creds: lnconnect.Credentials{},
			}

			err := optionFunc(mgr)

			if c.expectedError != nil {
				assert.EqualError(t, err, c.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mgr.creds.MacaroonBytes)
				assert.EqualValues(t, c.expectedManager.creds.MacaroonBytes, mgr.creds.MacaroonBytes)
			}
		})
	}
}
