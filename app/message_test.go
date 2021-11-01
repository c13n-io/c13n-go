package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
)

func TestVerifySignature(t *testing.T) {
	address := "000000000000000000000000000000000000000000000000000000000000000000"
	payload := []byte("payload")
	signature := []byte("dummy signature")

	cases := []struct {
		name             string
		sender           string
		msg              []byte
		sig              []byte
		mockCallResp     string
		mockCallErr      error
		expectedVerified bool
		expectedErr      error
	}{
		{
			name:             "verification error",
			sender:           address,
			msg:              payload,
			sig:              signature,
			mockCallResp:     "",
			mockCallErr:      fmt.Errorf("dummy error"),
			expectedVerified: false,
			expectedErr:      fmt.Errorf("dummy error"),
		},
		{
			name:             "signing address matches sender",
			sender:           address,
			msg:              payload,
			sig:              signature,
			mockCallResp:     address,
			mockCallErr:      nil,
			expectedVerified: true,
			expectedErr:      nil,
		},
		{
			name:             "signing address empty",
			sender:           address,
			msg:              payload,
			sig:              signature,
			mockCallResp:     "",
			mockCallErr:      nil,
			expectedVerified: false,
			expectedErr:      nil,
		},
		{
			name:             "sender address empty, signing address extracted",
			sender:           "",
			msg:              payload,
			sig:              signature,
			mockCallResp:     address,
			mockCallErr:      nil,
			expectedVerified: false,
			expectedErr:      nil,
		},
		{
			name:             "sender address does not match signing address",
			sender:           address,
			msg:              payload,
			sig:              signature,
			mockCallResp:     "111111111111111111111111111111111111111111111111111111111111111111",
			mockCallErr:      nil,
			expectedVerified: false,
			expectedErr:      nil,
		},
		{
			name:             "no message",
			sender:           address,
			msg:              nil,
			sig:              signature,
			expectedVerified: false,
			expectedErr:      nil,
		},
		{
			name:             "no signature",
			sender:           address,
			msg:              payload,
			sig:              nil,
			expectedVerified: false,
			expectedErr:      nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app := func() *App {
				lm := new(lnmock.LightManager)

				lm.On("VerifySignatureExtractPubkey", mock.Anything,
					c.msg, c.sig).Return(
					c.mockCallResp, c.mockCallErr).Once()

				return &App{
					LNManager: lm,
				}
			}()

			ctx := context.Background()
			ver, err := app.verifySignature(
				ctx, c.msg, c.sig, c.sender)

			switch c.expectedErr {
			case nil:
				assert.Equal(t, c.expectedVerified, ver)
				assert.NoError(t, err)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}
