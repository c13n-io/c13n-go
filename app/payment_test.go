package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/model"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

func TestSendPay(t *testing.T) {
	srcAddress := "000000000000000000000000000000000000000000000000000000000000000000"
	destAddress := "111111111111111111111111111111111111111111111111111111111111111111"

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: srcAddress,
		},
	}

	destNode, err := lnchat.NewNodeFromString(destAddress)
	assert.NoError(t, err)

	opts := model.MessageOptions{
		Anonymous: false,
	}

	paymentUpdatelist := []lnchat.PaymentUpdate{
		{
			Payment: &lnchat.Payment{
				Status:   lnchat.PaymentSUCCEEDED,
				Hash:     "1111111111111111111111111111111111111111111111111111111111111111",
				Preimage: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Htlcs: []lnchat.HTLCAttempt{
					{
						Status: lnrpc.HTLCAttempt_SUCCEEDED,
						Route: lnchat.Route{
							Amt:  lnchat.NewAmount(10000),
							Fees: lnchat.NewAmount(1000),
						},
					},
				},
			},
		},
	}

	preimageHash, err := lntypes.MakeHashFromStr(paymentUpdatelist[0].Payment.Hash)
	assert.NoError(t, err)

	preImage, err := lntypes.MakePreimageFromStr(paymentUpdatelist[0].Payment.Preimage)
	assert.NoError(t, err)

	cases := []struct {
		name            string
		discID          uint64
		discussion      *model.Discussion
		payReq          string
		decodedPayReq   *lnchat.PayReq
		amt             int64
		payload         string
		sendPaymentErr  error
		signMessageErr  error
		expectedErr     error
		expectedMessage *model.Message
	}{
		{
			name:           "Both Payreq and discussion ID provided",
			discID:         1,
			discussion:     nil,
			payReq:         "test payreq",
			decodedPayReq:  nil,
			amt:            0,
			payload:        "",
			sendPaymentErr: nil,
			expectedErr: errors.New("exactly one of payment request" +
				" and discussion must be specified"),
			expectedMessage: nil,
		},
		{
			name:            "Failed discussion retrieval",
			discID:          20,
			discussion:      nil,
			payReq:          "",
			decodedPayReq:   nil,
			amt:             0,
			payload:         "",
			sendPaymentErr:  nil,
			expectedErr:     errors.New("could not retrieve discussion: GetDiscussion: "),
			expectedMessage: nil,
		},
		{
			name:            "Decode Payreq failure",
			discID:          0,
			discussion:      nil,
			payReq:          "test payreq",
			decodedPayReq:   nil,
			amt:             0,
			payload:         "",
			sendPaymentErr:  nil,
			expectedErr:     errors.New("could not decode payment request: "),
			expectedMessage: nil,
		},
		{
			name:       "Retrieve discussion from Payreq failed",
			discID:     0,
			discussion: nil,
			payReq:     "test payreq",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
			},
			amt:             0,
			payload:         "",
			sendPaymentErr:  nil,
			expectedErr:     errors.New("could not retrieve discussion: retrieveOrCreateDiscussion: GetDiscussionByParticipants: "),
			expectedMessage: nil,
		},
		{
			name:   "SignMessage error",
			discID: 0,
			discussion: &model.Discussion{
				Options: model.MessageOptions{
					Anonymous: false,
				},
			},
			payReq: "test payreq",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
			},
			amt:             0,
			payload:         "",
			sendPaymentErr:  nil,
			signMessageErr:  errors.New(""),
			expectedErr:     errors.New("could not sign message payload: "),
			expectedMessage: nil,
		},
		{
			name:   "Success to dicsussion",
			discID: 1,
			discussion: &model.Discussion{
				ID: 1,
				Participants: []string{
					destAddress,
				},
				Options: DefaultOptions,
			},
			payReq:        "",
			decodedPayReq: nil,
			amt:           1000,
			expectedErr:   nil,
			payload:       "hello",
			expectedMessage: &model.Message{
				ID:             0,
				TotalFeesMsat:  1000,
				SenderVerified: true,
				Routes: []model.Route{
					{
						TotalTimeLock: 0,
						RouteAmtMsat:  10000,
						RouteFeesMsat: 1000,
						RouteHops:     []model.Hop{},
					},
				},
				PreimageHash: preimageHash[:],
				Preimage:     preImage,
				DiscussionID: 1,
				AmtMsat:      10000,
				Sender:       srcAddress,
				Receiver:     destAddress,
				Payload:      "hello",
			},
		},
		{
			name:   "Success to Payment request",
			discID: 0,
			discussion: &model.Discussion{
				ID:           1,
				Participants: []string{destAddress},
				Options:      DefaultOptions,
			},
			payReq: "test payreq",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
				Hash:        paymentUpdatelist[0].Payment.Hash,
			},
			amt:         10000,
			expectedErr: nil,
			payload:     "hello",
			expectedMessage: &model.Message{
				ID:             0,
				TotalFeesMsat:  1000,
				SenderVerified: true,
				Routes: []model.Route{
					{
						TotalTimeLock: 0,
						RouteAmtMsat:  10000,
						RouteFeesMsat: 1000,
						RouteHops:     []model.Hop{},
					},
				},
				PreimageHash: preimageHash[:],
				Preimage:     preImage,
				DiscussionID: 1,
				AmtMsat:      10000,
				Sender:       srcAddress,
				Receiver:     destAddress,
				Payload:      "hello",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(
					uint64(1), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates",
					mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

				if c.decodedPayReq == nil {
					mockLNManager.On("DecodePayReq", mock.Anything, c.payReq).Return(c.decodedPayReq, errors.New(""))
				} else {
					mockLNManager.On("DecodePayReq", mock.Anything, c.payReq).Return(c.decodedPayReq, nil)
					if c.discussion == nil {
						mockDB.On("GetDiscussionByParticipants", []string{destAddress}).Return(c.discussion, errors.New("")).Once()
					} else {
						mockDB.On("GetDiscussionByParticipants", []string{destAddress}).Return(c.discussion, nil).Once()
					}
				}

				mockLNManager.On("SignMessage", mock.Anything, mock.Anything).Return([]byte("sig"), c.signMessageErr).Once()

				var recipients []string
				if c.discussion != nil {
					recipients = c.discussion.Participants
				} else {
					recipients = []string{destAddress}
				}

				for _, recipient := range recipients {
					paymentUpdates := func() <-chan lnchat.PaymentUpdate {
						ch := make(chan lnchat.PaymentUpdate)

						go func(c chan lnchat.PaymentUpdate) {
							defer close(c)

							for _, update := range paymentUpdatelist {
								c <- update
							}
						}(ch)
						return ch
					}()

					mockLNManager.On("SendPayment", mock.Anything, recipient, lnchat.NewAmount(c.amt), c.payReq, lnchat.PaymentOptions{
						FeeLimitMsat:   3000,
						FinalCltvDelta: 20,
						TimeoutSecs:    30,
					}, mock.Anything, mock.Anything).Return(paymentUpdates, c.sendPaymentErr).Once()
				}

				mockDB.On("AddPayments", mock.AnythingOfType("*model.Payment")).Return(
					nil).Once()

				mockDB.On("AddRawMessage", mock.AnythingOfType("*model.RawMessage")).Return(
					nil).Once()

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				if c.discussion != nil {
					mockDB.On("GetDiscussion", c.discID).Return(
						c.discussion, nil).Once()
				} else {
					mockDB.On("GetDiscussion", c.discID).Return(
						c.discussion, errors.New("")).Once()
				}

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			msg, err := app.SendPay(ctxt, c.payload, c.amt, c.discID, c.payReq, opts)

			if c.expectedErr != nil {
				assert.Nil(t, msg)
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expectedMessage, msg)
			}
		})
	}
}
