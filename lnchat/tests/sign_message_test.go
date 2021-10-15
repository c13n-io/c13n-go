package itest

import (
	"context"
	"errors"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"
	"github.com/tv42/zbase32"
)

func testSignMessage(net *lntest.NetworkHarness, t *harnessTest) {
	cases := []struct {
		name        string
		message     []byte
		expectedErr error
	}{
		{
			name:        "Successful signing",
			message:     []byte("test_message"),
			expectedErr: nil,
		},
		{
			name:        "Empty message",
			message:     []byte(""),
			expectedErr: errors.New(""),
		},
	}

	// Create managers
	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	mgrBob, err := createNodeManager(net.Bob)
	assert.NoError(t.t, err)

	for _, c := range cases {
		t.t.Run(c.name, func(subTest *testing.T) {
			ctxb := context.Background()
			sig, err := mgrAlice.SignMessage(ctxb, c.message)

			if c.expectedErr == nil {
				assert.NotNil(t.t, sig)
				assert.NoError(t.t, err)

				// Verify Message
				req := &lnrpc.VerifyMessageRequest{
					Msg:       c.message,
					Signature: zbase32.EncodeToString(sig),
				}

				resp, err := net.Alice.VerifyMessage(ctxb, req)
				assert.NoError(t.t, err, "Message verification failed")

				assert.Equal(t.t, net.Alice.PubKeyStr, resp.Pubkey)
			} else {
				assert.Nil(t.t, sig)
				assert.Error(t.t, err)
			}
		})
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)

	err = mgrBob.Close()
	assert.NoError(t.t, err)
}
