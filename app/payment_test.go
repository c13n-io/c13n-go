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

	paymentOpts := lnchat.PaymentOptions{
		FeeLimitMsat:   3000,
		FinalCltvDelta: 20,
		TimeoutSecs:    30,
	}

	opts := model.MessageOptions{
		Anonymous: false,
	}

	paymentUpdateList := []lnchat.PaymentUpdate{
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

	preimageHash, err := lntypes.MakeHashFromStr(paymentUpdateList[0].Payment.Hash)
	assert.NoError(t, err)

	preImage, err := lntypes.MakePreimageFromStr(paymentUpdateList[0].Payment.Preimage)
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
				Hash:        paymentUpdateList[0].Payment.Hash,
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

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(1), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)
				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(nil, nil)

				if c.decodedPayReq == nil {
					mockLNManager.On("DecodePayReq", mock.Anything, c.payReq).Return(
						c.decodedPayReq, errors.New(""))
				} else {
					mockLNManager.On("DecodePayReq", mock.Anything, c.payReq).Return(
						c.decodedPayReq, nil)
					if c.discussion == nil {
						mockDB.On("GetDiscussionByParticipants",
							[]string{destAddress}).Return(
							c.discussion, errors.New("")).Once()
					} else {
						mockDB.On("GetDiscussionByParticipants",
							[]string{destAddress}).Return(
							c.discussion, nil).Once()
					}
				}

				mockLNManager.On("SignMessage", mock.Anything, mock.Anything).Return(
					[]byte("sig"), c.signMessageErr).Once()

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

							for _, update := range paymentUpdateList {
								c <- update
							}
						}(ch)
						return ch
					}()

					mockLNManager.On("SendPayment", mock.Anything, recipient,
						lnchat.NewAmount(c.amt), c.payReq, paymentOpts, mock.Anything,
						mock.Anything).Return(paymentUpdates, c.sendPaymentErr).Once()
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

func TestSendPayment(t *testing.T) {
	srcAddress := "000000000000000000000000000000000000000000000000000000000000000000"
	destAddress := "111111111111111111111111111111111111111111111111111111111111111111"

	dummyHash := "abcdxdcba"
	dummyPreimage := "17422471"

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: srcAddress,
		},
	}

	value := int64(15000)
	fee := int64(500)

	opts := lnchat.PaymentOptions{
		FeeLimitMsat:   5000,
		FinalCltvDelta: 20,
		TimeoutSecs:    100,
	}

	payment := &model.Payment{
		PayerAddress: srcAddress,
		PayeeAddress: destAddress,
		Payment: lnchat.Payment{
			Hash:           dummyHash,
			Preimage:       dummyPreimage,
			Value:          lnchat.NewAmount(value),
			PaymentRequest: "dog",
			Status:         lnchat.PaymentSUCCEEDED,
			PaymentIndex:   0,
			Htlcs: []lnchat.HTLCAttempt{
				{
					Status: lnrpc.HTLCAttempt_SUCCEEDED,
					Route: lnchat.Route{
						Amt:  lnchat.NewAmount(value),
						Fees: lnchat.NewAmount(fee),
					},
				},
			},
		},
	}

	paymentUpdateList := []lnchat.PaymentUpdate{
		{
			Payment: &lnchat.Payment{
				Hash:           dummyHash,
				Preimage:       dummyPreimage,
				Value:          lnchat.NewAmount(value),
				PaymentRequest: "dog",
				Status:         lnchat.PaymentSUCCEEDED,
				PaymentIndex:   0,
				Htlcs: []lnchat.HTLCAttempt{
					{
						Status: lnrpc.HTLCAttempt_SUCCEEDED,
						Route: lnchat.Route{
							Amt:  lnchat.NewAmount(value),
							Fees: lnchat.NewAmount(fee),
						},
					},
				},
			},
		},
	}

	destNode, _ := lnchat.NewNodeFromString(destAddress)

	cases := []struct {
		name            string
		dest            string
		payReq          string
		decodedPayReq   *lnchat.PayReq
		amt             int64
		tlvs            map[uint64][]byte
		sendPaymentErr  error
		signMessageErr  error
		expectedErr     error
		expectedPayment *model.Payment
	}{
		{
			name:   "Both Payreq and address provided",
			dest:   "111111111111111111111111111111111111111111111111111111111111111111",
			payReq: "dummyPayreq",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
			},
			amt:            0,
			sendPaymentErr: nil,
			expectedErr: errors.New("exactly one of payment request" +
				" and destination address must be specified"),
			expectedPayment: nil,
		},
		{
			name:   "Payreq provided",
			dest:   "",
			payReq: "dummyPayreq",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
			},
			amt:             value,
			sendPaymentErr:  nil,
			expectedErr:     nil,
			expectedPayment: payment,
		},
		{
			name:   "Address provided",
			dest:   "111111111111111111111111111111111111111111111111111111111111111111",
			payReq: "",
			decodedPayReq: &lnchat.PayReq{
				Destination: destNode,
			},
			amt:             value,
			sendPaymentErr:  nil,
			expectedErr:     nil,
			expectedPayment: payment,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(1), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)
				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(nil, nil)

				mockLNManager.On("DecodePayReq", mock.Anything, c.payReq).Return(
					c.decodedPayReq, nil)

				mockLNManager.On("SignMessage", mock.Anything, mock.Anything).Return(
					[]byte("sig"), c.signMessageErr).Once()

				paymentUpdates := func() <-chan lnchat.PaymentUpdate {
					ch := make(chan lnchat.PaymentUpdate)

					go func(c chan lnchat.PaymentUpdate) {
						defer close(c)

						for _, update := range paymentUpdateList {
							c <- update
						}
					}(ch)
					return ch
				}()

				mockLNManager.On("SendPayment", mock.Anything, c.dest, lnchat.NewAmount(c.amt),
					c.payReq, opts, c.tlvs, mock.Anything).Return(
					paymentUpdates, c.sendPaymentErr).Once()

				mockDB.On("AddPayments", mock.AnythingOfType("*model.Payment")).Return(nil).Once()

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

			payment, err := app.SendPayment(ctxt, c.dest, c.amt, c.payReq, opts, nil)
			if c.expectedErr != nil {
				assert.Nil(t, payment)
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expectedPayment, payment)
			}
		})
	}
}
