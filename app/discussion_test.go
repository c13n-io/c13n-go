package app

import (
	"context"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/model"
	"github.com/c13n-io/c13n-go/store"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

func TestAddDiscussion(t *testing.T) {
	selfAddress := "000000000000000000000000000000000000000000000000000000000000000000"
	participantAddress := "111111111111111111111111111111111111111111111111111111111111111111"
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: selfAddress,
		},
	}

	cases := []struct {
		name                      string
		discussionToAdd           *model.Discussion
		expectedAddDiscussionResp *model.Discussion
		expectedAddDiscussionErr  error
		expectedErr               error
	}{
		{
			name: "Success",
			discussionToAdd: &model.Discussion{
				Participants: []string{participantAddress},
				Options: model.MessageOptions{
					FeeLimitMsat: 1000,
					Anonymous:    false,
				},
			},
			expectedAddDiscussionResp: &model.Discussion{
				ID:           1,
				Participants: []string{participantAddress},
				Options: model.MessageOptions{
					FeeLimitMsat: 1000,
					Anonymous:    false,
				},
			},
			expectedAddDiscussionErr: nil,
			expectedErr:              nil,
		},
		{
			name: "Discussion without fee limit - use default",
			discussionToAdd: &model.Discussion{
				Participants: []string{participantAddress},
			},
			expectedAddDiscussionResp: &model.Discussion{
				ID:           1,
				Participants: []string{participantAddress},
				Options: model.MessageOptions{
					FeeLimitMsat: DefaultOptions.FeeLimitMsat,
					Anonymous:    false,
				},
			},
			expectedAddDiscussionErr: nil,
			expectedErr:              nil,
		},
		{
			name: "Discussion already exists",
			discussionToAdd: &model.Discussion{
				Participants: []string{participantAddress},
				Options: model.MessageOptions{
					FeeLimitMsat: 1200,
				},
			},
			expectedAddDiscussionResp: nil,
			expectedAddDiscussionErr:  store.ErrDiscussionAlreadyExists,
			expectedErr: Error{
				Kind:    DiscussionAlreadyExists,
				Err:     store.ErrDiscussionAlreadyExists,
				details: "AddDiscussion",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager,
				mockDB *dbmock.Database) (*lnmock.LightManager,
				*dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(42), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)
				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(42),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(nil, nil)

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockDB.On("AddDiscussion", c.discussionToAdd).Return(
					c.expectedAddDiscussionResp, c.expectedAddDiscussionErr).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			discussion, err := app.AddDiscussion(ctxt, c.discussionToAdd)

			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.EqualValues(t, c.expectedAddDiscussionResp, discussion)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, discussion)
			}
		})
	}
}

func TestRemoveDiscussion(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	cases := []struct {
		name                 string
		idToRemove           uint64
		removeDiscussionResp *model.Discussion
		removeDiscussionErr  error
		expectedErr          error
	}{
		{
			name:       "Success",
			idToRemove: 1,
			removeDiscussionResp: &model.Discussion{
				ID: 41,
				Participants: []string{
					"111111111111111111111111111111111111111111111111111111111111111111",
				},
				Options: model.MessageOptions{
					FeeLimitMsat: 32000,
					Anonymous:    false,
				},
			},
			removeDiscussionErr: nil,
			expectedErr:         Error{},
		},
		{
			name:                 "Not found",
			idToRemove:           10,
			removeDiscussionResp: nil,
			removeDiscussionErr:  store.ErrDiscussionNotFound,
			expectedErr: Error{
				Kind:    DiscussionNotFound,
				details: "RemoveDiscussion",
				Err:     store.ErrDiscussionNotFound,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(42), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)
				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(42),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(nil, nil)

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockDB.On("RemoveDiscussion", c.idToRemove).Return(
					c.removeDiscussionResp, c.removeDiscussionErr).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			err := app.RemoveDiscussion(ctxt, c.idToRemove)

			switch c.expectedErr {
			case Error{}:
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}

func TestGetDiscussionStatistics(t *testing.T) {
	selfAddress := "000000000000000000000000000000000000000000000000000000000000000000"
	correspondentAddress := "111111111111111111111111111111111111111111111111111111111111111111"

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: selfAddress,
		},
	}

	var discussionID uint64 = 32
	msgHistory := []model.MessageAggregate{
		{
			RawMessage: &model.RawMessage{
				ID:                39,
				DiscussionID:      discussionID,
				RawPayload:        []byte("third message payload"),
				Sender:            selfAddress,
				Signature:         []byte(selfAddress),
				SignatureVerified: true,
				PaymentIndexes:    []uint64{0},
			},
			Payments: []*model.Payment{
				{
					PayerAddress: selfAddress,
					PayeeAddress: correspondentAddress,
					Payment: lnchat.Payment{
						Status: lnchat.PaymentSUCCEEDED,
						Htlcs: []lnchat.HTLCAttempt{
							{
								Status: lnrpc.HTLCAttempt_SUCCEEDED,
								Route: lnchat.Route{
									Amt:  lnchat.NewAmount(21000),
									Fees: lnchat.NewAmount(2003),
								},
							},
						},
					},
				},
			},
		},
		{
			RawMessage: &model.RawMessage{
				ID:                 31,
				DiscussionID:       discussionID,
				RawPayload:         []byte("second message payload"),
				Sender:             correspondentAddress,
				Signature:          []byte(correspondentAddress),
				SignatureVerified:  true,
				InvoiceSettleIndex: 1,
			},
			Invoice: &model.Invoice{
				Invoice: lnchat.Invoice{
					AmtPaid: 5310,
				},
			},
		},
		{
			RawMessage: &model.RawMessage{
				ID:                23,
				DiscussionID:      discussionID,
				RawPayload:        []byte("first message payload"),
				Sender:            selfAddress,
				Signature:         []byte(selfAddress),
				SignatureVerified: true,
				PaymentIndexes:    []uint64{0},
			},
			Payments: []*model.Payment{
				{
					PayerAddress: selfAddress,
					PayeeAddress: correspondentAddress,
					Payment: lnchat.Payment{
						Status: lnchat.PaymentSUCCEEDED,
						Htlcs: []lnchat.HTLCAttempt{
							{
								Status: lnrpc.HTLCAttempt_SUCCEEDED,
								Route: lnchat.Route{
									Amt:  lnchat.NewAmount(12001),
									Fees: lnchat.NewAmount(1200),
								},
							},
						},
					},
				},
			},
		},
	}

	historyStatistics := model.DiscussionStatistics{
		AmtMsatSent:      33001,
		AmtMsatFees:      3203,
		AmtMsatReceived:  5310,
		MessagesSent:     2,
		MessagesReceived: 1,
	}

	_ = msgHistory
	_ = selfInfo

	cases := []struct {
		name                     string
		getDiscussionHistoryResp []model.MessageAggregate
		getDiscussionHistoryErr  error
		expectedResponse         *model.DiscussionStatistics
		expectedErr              error
	}{
		{
			name:                     "Success",
			getDiscussionHistoryResp: msgHistory,
			getDiscussionHistoryErr:  nil,
			expectedResponse:         &historyStatistics,
			expectedErr:              nil,
		},
		{
			name:                     "Discussion not found",
			getDiscussionHistoryResp: nil,
			getDiscussionHistoryErr:  store.ErrDiscussionNotFound,
			expectedResponse:         nil,
			expectedErr: Error{
				Kind:    DiscussionNotFound,
				details: "could not retrieve discussion messages",
				Err:     store.ErrDiscussionNotFound,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager,
				mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(42), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)
				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(42),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(nil, nil)

				mockDB.On("GetMessages",
					discussionID, model.PageOptions{}).Return(
					c.getDiscussionHistoryResp, c.getDiscussionHistoryErr).Once()

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxc, cancel := context.WithCancel(context.Background())
			defer cancel()

			stats, err := app.GetDiscussionStatistics(ctxc, discussionID)

			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.EqualValues(t, c.expectedResponse, stats)
			default:
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, stats)
			}
		})
	}

}
