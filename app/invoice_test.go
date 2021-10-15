package app

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-backend/lnchat"
	lnmock "github.com/c13n-io/c13n-backend/lnchat/mocks"
	"github.com/c13n-io/c13n-backend/model"
	dbmock "github.com/c13n-io/c13n-backend/store/mocks"
)

func mustJsonMarshalMessage(t *testing.T, participants []string, payload string) []byte {
	type compositePayload struct {
		Participants []string `json:"participants"`
		Message      string   `json:"message"`
	}

	data, err := json.Marshal(compositePayload{
		Participants: participants,
		Message:      payload,
	})
	require.NoError(t, err)

	return data
}

func mustExtractRawMessage(t *testing.T, inv *lnchat.Invoice) *model.RawMessage {
	rawMsg, err := payloadExtractor(inv, func(_, _ []byte, _ string) (
		bool, error) {
		return true, nil
	})
	require.NoError(t, err)

	return rawMsg
}

func mustMakePreimageHash(t *testing.T, hash string) []byte {
	h, err := lntypes.MakeHashFromStr(hash)
	require.NoError(t, err)

	return h[:]
}

func mustMakePreimage(t *testing.T, preimage []byte) lntypes.Preimage {
	p, err := lntypes.MakePreimage(preimage)
	require.NoError(t, err)

	return p
}

func TestSubscribeInvoices(t *testing.T) {
	selfAddr, err := lnchat.NewNodeFromString(
		"111111111111111111111111111111111111111111111111111111111111111111")
	require.NoError(t, err)

	srcAddr, err := lnchat.NewNodeFromString(
		"000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: selfAddr.String(),
		},
	}

	invoiceUpdateList := []lnchat.InvoiceUpdate{
		{
			Inv: &lnchat.Invoice{
				Hash:           "0000000000000000000000000000000000000000000000000000000000000000",
				Preimage:       []byte("00000000000000000000000000000000"),
				PaymentRequest: "dummy payment request",
				Value:          lnchat.NewAmount(123),
				AmtPaid:        lnchat.NewAmount(124),
				CreatedTimeSec: time.Now().Unix(),
				SettleTimeSec:  time.Now().Add(3 * time.Second).Unix(),
				State:          lnchat.InvoiceSETTLED,
				SettleIndex:    738,
				Htlcs: []lnchat.InvoiceHTLC{
					{
						State: lnrpc.InvoiceHTLCState_SETTLED,
						CustomRecords: map[uint64][]byte{
							PayloadTypeKey: mustJsonMarshalMessage(t,
								[]string{selfAddr.String()}, "test message"),
							SenderTypeKey:    srcAddr.Bytes(),
							SignatureTypeKey: []byte("a dummy signature"),
						},
					},
				},
			},
			Err: nil,
		},
		{
			Inv: &lnchat.Invoice{
				Hash:           "0000000000000000000000000000000000000000000000000000000000000000",
				Preimage:       []byte("00000000000000000000000000000001"),
				PaymentRequest: "another dummy payreq",
				Value:          lnchat.NewAmount(400),
				AmtPaid:        lnchat.NewAmount(402),
				CreatedTimeSec: time.Now().Unix(),
				SettleTimeSec:  time.Now().Add(4 * time.Second).Unix(),
				State:          lnchat.InvoiceSETTLED,
				SettleIndex:    741,
				Htlcs: []lnchat.InvoiceHTLC{
					{
						State: lnrpc.InvoiceHTLCState_SETTLED,
					},
				},
			},
			Err: nil,
		},
		{
			Inv: &lnchat.Invoice{
				Hash:           "0000000000000000000000000000000000000000000000000000000000000000",
				Preimage:       []byte("00000000000000000000000000000002"),
				PaymentRequest: "paying HTLC has unexpected payload structure",
				Value:          lnchat.NewAmount(318),
				AmtPaid:        lnchat.NewAmount(801),
				CreatedTimeSec: time.Now().Unix(),
				SettleTimeSec:  time.Now().Add(3 * time.Second).Unix(),
				State:          lnchat.InvoiceSETTLED,
				SettleIndex:    753,
				Htlcs: []lnchat.InvoiceHTLC{
					{
						State: lnrpc.InvoiceHTLCState_SETTLED,
						CustomRecords: map[uint64][]byte{
							PayloadTypeKey:   []byte("unexpected payload structure"),
							SenderTypeKey:    srcAddr.Bytes(),
							SignatureTypeKey: []byte("a dummy signature"),
						},
					},
				},
			},
			Err: nil,
		},
	}

	type invoiceUpdateOp struct {
		data                     lnchat.InvoiceUpdate
		addInvoiceErr            error
		payloadExists            bool
		payloadSigned            bool
		verifySigExtractedPubkey string
		verifySigErr             error
		rawMsg                   *model.RawMessage
		canUnmarshalPayload      bool
		discExists               bool
		discParticipants         []string
		discussion               *model.Discussion
		discID                   uint64
		getDiscByParticipantsErr error
		addDiscussionErr         error
		addRawMsgErr             error
		addRawMsgID              uint64
		message                  *model.Message
	}

	cases := []struct {
		name                string
		subscrInvUpdatesErr error
		invoiceUpdateOps    []invoiceUpdateOp
	}{
		{
			name:                "Success",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps: []invoiceUpdateOp{
				{
					data:                     invoiceUpdateList[0],
					addInvoiceErr:            nil,
					payloadExists:            true,
					payloadSigned:            true,
					verifySigExtractedPubkey: srcAddr.String(),
					verifySigErr:             nil,
					rawMsg:                   mustExtractRawMessage(t, invoiceUpdateList[0].Inv),
					canUnmarshalPayload:      true,
					discExists:               true,
					discParticipants:         []string{srcAddr.String()},
					discussion: &model.Discussion{
						Participants:  []string{srcAddr.String()},
						LastReadID:    0,
						LastMessageID: 33,
						Options:       DefaultOptions,
					},
					discID:                   13,
					getDiscByParticipantsErr: nil,
					addDiscussionErr:         nil,
					addRawMsgErr:             nil,
					addRawMsgID:              42,
					message: &model.Message{
						ID:             42,
						DiscussionID:   13,
						Payload:        "test message",
						AmtMsat:        124,
						Sender:         srcAddr.String(),
						Receiver:       selfAddr.String(),
						SenderVerified: true,
						SentTimeNs:     time.Unix(invoiceUpdateList[0].Inv.CreatedTimeSec, 0).UnixNano(),
						ReceivedTimeNs: time.Unix(invoiceUpdateList[0].Inv.SettleTimeSec, 0).UnixNano(),
						Index:          invoiceUpdateList[0].Inv.SettleIndex,
						PayReq:         "dummy payment request",
						SuccessProb:    1.,
						PreimageHash:   mustMakePreimageHash(t, invoiceUpdateList[0].Inv.Hash),
						Preimage:       mustMakePreimage(t, invoiceUpdateList[0].Inv.Preimage),
					},
				},
			},
		},
		{
			name:                "Subscription terminated",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps:    []invoiceUpdateOp{},
		},
		{
			name:                "AddInvoice error",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps: []invoiceUpdateOp{
				{
					data:          invoiceUpdateList[1],
					addInvoiceErr: fmt.Errorf("AddInvoice dummy error"),
					payloadExists: false,
				},
			},
		},
		{
			name:                "Discussion retrieval error",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps: []invoiceUpdateOp{
				{
					data:                     invoiceUpdateList[0],
					addInvoiceErr:            nil,
					payloadExists:            true,
					payloadSigned:            true,
					verifySigExtractedPubkey: srcAddr.String(),
					verifySigErr:             nil,
					rawMsg:                   mustExtractRawMessage(t, invoiceUpdateList[0].Inv),
					canUnmarshalPayload:      true,
					discExists:               true,
					discParticipants:         []string{srcAddr.String()},
					discussion:               nil,
					discID:                   0,
					getDiscByParticipantsErr: fmt.Errorf("dummy GetDiscussionByParticipants error"),
					addDiscussionErr:         nil,
				},
			},
		},
		{
			name:                "AddRawMessage error",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps: []invoiceUpdateOp{
				{
					data:                     invoiceUpdateList[0],
					addInvoiceErr:            nil,
					payloadExists:            true,
					payloadSigned:            true,
					verifySigExtractedPubkey: srcAddr.String(),
					verifySigErr:             nil,
					rawMsg:                   mustExtractRawMessage(t, invoiceUpdateList[0].Inv),
					canUnmarshalPayload:      true,
					discExists:               true,
					discParticipants:         []string{srcAddr.String()},
					discussion: &model.Discussion{
						Participants:  []string{srcAddr.String()},
						LastReadID:    0,
						LastMessageID: 33,
						Options:       DefaultOptions,
					},
					discID:                   13,
					getDiscByParticipantsErr: nil,
					addDiscussionErr:         nil,
					addRawMsgErr:             fmt.Errorf("dummy AddRawMessage error"),
					addRawMsgID:              0,
					message:                  nil,
				},
			},
		},
		{
			name:                "Unexpected payload structure",
			subscrInvUpdatesErr: nil,
			invoiceUpdateOps: []invoiceUpdateOp{
				{
					data:                     invoiceUpdateList[2],
					addInvoiceErr:            nil,
					payloadExists:            true,
					payloadSigned:            true,
					verifySigExtractedPubkey: srcAddr.String(),
					verifySigErr:             nil,
					rawMsg:                   mustExtractRawMessage(t, invoiceUpdateList[2].Inv),
					canUnmarshalPayload:      false,
				},
			},
		},
	}

	discWithID := func(disc *model.Discussion, id uint64) *model.Discussion {
		if disc == nil {
			return nil
		}
		ret := *disc
		ret.ID = id
		return &ret
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {

			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				testcaseTermination := make(chan bool)

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(
					uint64(1), nil).Once()

				// Mock invoice update subscription channel
				invoiceUpdateCh := func() <-chan lnchat.InvoiceUpdate {
					ch := make(chan lnchat.InvoiceUpdate)

					go func(channel chan lnchat.InvoiceUpdate) {
						defer close(channel)

						for _, update := range c.invoiceUpdateOps {
							channel <- update.data
						}

						<-testcaseTermination
					}(ch)
					return ch
				}()

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(
					invoiceUpdateCh, c.subscrInvUpdatesErr).Once()

				for _, invUpdate := range c.invoiceUpdateOps {
					if invUpdate.data.Inv != nil {
						invModel := &model.Invoice{
							CreatorAddress: selfAddr.String(),
							Invoice:        *invUpdate.data.Inv,
						}
						mockDB.On("AddInvoice", invModel).Return(
							invUpdate.addInvoiceErr).Once()
					}
					if !invUpdate.payloadExists {
						continue
					}
					if invUpdate.payloadSigned {
						payloadMap := invUpdate.data.Inv.Htlcs[0].CustomRecords
						msg, sig := payloadMap[PayloadTypeKey], payloadMap[SignatureTypeKey]

						mockLNManager.On("VerifySignatureExtractPubkey", mock.Anything,
							msg, sig).Return(invUpdate.verifySigExtractedPubkey,
							invUpdate.verifySigErr).Once()
					}
					if !invUpdate.canUnmarshalPayload {
						continue
					}
					if invUpdate.canUnmarshalPayload {
						disc := discWithID(invUpdate.discussion, invUpdate.discID)

						mockDB.On("GetDiscussionByParticipants",
							invUpdate.discParticipants).Return(disc,
							invUpdate.getDiscByParticipantsErr).Once()
					}
					if !invUpdate.discExists {
						disc := discWithID(invUpdate.discussion, invUpdate.discID)

						mockDB.On("AddDiscussion", invUpdate.discussion).Return(
							disc, invUpdate.addDiscussionErr).Once()
					}
					if invUpdate.rawMsg != nil && invUpdate.discussion != nil {
						rawMsg := invUpdate.rawMsg
						rawMsg.DiscussionID = invUpdate.discID

						mockDB.On("AddRawMessage", rawMsg).Return(invUpdate.addRawMsgErr).Run(
							func(args mock.Arguments) {
								arg := args.Get(0).(*model.RawMessage)
								arg.ID = invUpdate.addRawMsgID
							}).Once()
					}
				}

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockStopFunc := func() {
					close(testcaseTermination)
				}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxc, cancel := context.WithCancel(context.Background())
			defer cancel()

			msgCh, err := app.SubscribeMessages(ctxc)
			assert.NoError(t, err)

			for i := 0; i < len(c.invoiceUpdateOps); i++ {
				if !c.invoiceUpdateOps[i].payloadExists ||
					!c.invoiceUpdateOps[i].canUnmarshalPayload ||
					c.invoiceUpdateOps[i].message == nil {
					continue
				}

				pubMsg, ok := <-msgCh
				assert.Truef(t, ok, "Expected message %d not received prior to channel close", i)
				var msg model.Message
				err := json.Unmarshal(pubMsg.Payload, &msg)
				assert.NoError(t, err)

				assert.EqualValues(t, c.invoiceUpdateOps[i].message, &msg)
			}
		})
	}
}
