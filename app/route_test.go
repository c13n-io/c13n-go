package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/model"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

func payOptsWithFeeLimit(feeLimit int64) lnchat.PaymentOptions {
	return DefaultOptions.WithFeeLimit(feeLimit).GetPaymentOptions()
}

func mustCreatePayload(t *testing.T, participants []string,
	payload, sender string, signature []byte) map[uint64][]byte {

	rawPayload := mustJsonMarshalMessage(t, participants, payload)

	raw := &model.RawMessage{
		RawPayload: rawPayload,
		Sender:     sender,
		Signature:  signature,
	}

	return marshalPayload(raw)
}

func TestEstimatePayment(t *testing.T) {
	srcAddress := "000000000000000000000000000000000000000000000000000000000000000000"
	destAddress := "111111111111111111111111111111111111111111111111111111111111111111"

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: srcAddress,
		},
	}

	destNode, err := lnchat.NewNodeFromString(destAddress)
	require.NoError(t, err)

	type getRouteCall struct {
		recipient     string
		amt           int64
		payOpts       lnchat.PaymentOptions
		payload       map[uint64][]byte
		expectedRoute *lnchat.Route
		expectedProb  float64
		expectedErr   error
	}

	defaultTestOpts := model.MessageOptions{
		FeeLimitMsat: 3200,
		Anonymous:    false,
	}

	testPayload := "test payload"

	zeroHash, zeroPreimage := lntypes.ZeroHash, lntypes.Preimage{}

	discussions := []model.Discussion{
		model.Discussion{
			ID: 42,
			Participants: []string{
				destAddress,
			},
			Options: DefaultOptions,
		},
	}

	cases := []struct {
		name             string
		discID           uint64
		amt              int64
		payload          string
		opts             model.MessageOptions
		discussion       *model.Discussion
		getDiscussionErr error
		signature        []byte
		signMessageErr   error
		getRouteCalls    []getRouteCall
		expectedMessage  *model.Message
		expectedErr      error
	}{
		{
			name:             "Success",
			discID:           discussions[0].ID,
			amt:              1023,
			payload:          testPayload,
			opts:             defaultTestOpts,
			discussion:       &discussions[0],
			getDiscussionErr: nil,
			signature:        []byte("dummy signature"),
			signMessageErr:   nil,
			getRouteCalls: []getRouteCall{
				{
					recipient: discussions[0].Participants[0],
					amt:       1023,
					payOpts:   payOptsWithFeeLimit(3200),
					payload: mustCreatePayload(t, discussions[0].Participants,
						testPayload, srcAddress, []byte("dummy signature")),
					expectedRoute: &lnchat.Route{
						TimeLock: 321,
						Amt:      lnchat.NewAmount(1023),
						Hops: []lnchat.RouteHop{
							{
								ChannelID:    0x01,
								NodeID:       destNode,
								AmtToForward: lnchat.NewAmount(1023),
								Expiry:       333,
							},
						},
					},
					expectedProb: .75,
					expectedErr:  nil,
				},
			},
			expectedMessage: &model.Message{
				DiscussionID:   discussions[0].ID,
				Payload:        testPayload,
				AmtMsat:        1023,
				Sender:         srcAddress,
				Receiver:       destAddress,
				SenderVerified: true,
				TotalFeesMsat:  0,
				Routes: []model.Route{
					model.Route{
						TotalTimeLock: 321,
						RouteAmtMsat:  1023,
						RouteFeesMsat: 0,
						RouteHops: []model.Hop{
							model.Hop{
								ChanID:           0x01,
								HopAddress:       destNode.String(),
								AmtToForwardMsat: 1023,
								FeeMsat:          0,
							},
						},
					},
				},
				PreimageHash: zeroHash[:],
				Preimage:     zeroPreimage,
				SuccessProb:  .75,
			},
			expectedErr: nil,
		},
		{
			name:             "GetDiscussion error",
			discID:           41,
			amt:              101,
			payload:          "test should fail to find discussion",
			opts:             defaultTestOpts,
			discussion:       nil,
			getDiscussionErr: fmt.Errorf("some error"),
			expectedMessage:  nil,
			expectedErr:      fmt.Errorf("could not retrieve discussion: GetDiscussion: some error"),
		},
		{
			name:             "SignMessage error",
			discID:           42,
			amt:              102,
			payload:          "test should fail to sign payload",
			opts:             defaultTestOpts,
			discussion:       &discussions[0],
			getDiscussionErr: nil,
			signature:        nil,
			signMessageErr:   fmt.Errorf("some error"),
			expectedMessage:  nil,
			expectedErr:      fmt.Errorf("could not sign message payload: some error"),
		},
		{
			name:             "GetRoute error",
			discID:           42,
			amt:              103,
			payload:          "test should fail to find route",
			opts:             defaultTestOpts,
			discussion:       &discussions[0],
			getDiscussionErr: nil,
			signature:        []byte("dummy signature"),
			signMessageErr:   nil,
			getRouteCalls: []getRouteCall{
				{
					recipient: discussions[0].Participants[0],
					amt:       103,
					payOpts:   payOptsWithFeeLimit(3200),
					payload: mustCreatePayload(t, discussions[0].Participants,
						"test should fail to find route", srcAddress, []byte("dummy signature")),
					expectedRoute: nil,
					expectedProb:  .0,
					expectedErr:   fmt.Errorf("some error"),
				},
			},
			expectedMessage: nil,
			expectedErr:     fmt.Errorf("could not find route to %s: some error", destAddress),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(0), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(0),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

				mockDB.On("GetDiscussion", c.discID).Return(
					c.discussion, c.getDiscussionErr).Once()

				if c.getDiscussionErr == nil {
					if !c.opts.Anonymous {
						marshalledPayload := mustJsonMarshalMessage(t,
							c.discussion.Participants, c.payload)
						mockLNManager.On("SignMessage", mock.Anything, marshalledPayload).Return(
							c.signature, c.signMessageErr).Once()
					}

					for _, call := range c.getRouteCalls {
						mockLNManager.On("GetRoute", mock.Anything,
							call.recipient, lnchat.NewAmount(call.amt),
							call.payOpts, call.payload).Return(
							call.expectedRoute, call.expectedProb, call.expectedErr).Once()
					}
				}

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			msg, err := app.EstimatePayment(ctxt, c.payload, c.amt, c.discID, c.opts)

			switch c.expectedErr {
			case nil:
				msg.SentTimeNs, msg.ReceivedTimeNs = 0, 0

				assert.NoError(t, err)
				assert.Equal(t, c.expectedMessage, msg)
			default:
				assert.Nil(t, msg)
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}
