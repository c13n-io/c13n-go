package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/model"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

func TestSubscribePayments(t *testing.T) {
	selfAddr, err := lnchat.NewNodeFromString(
		"111111111111111111111111111111111111111111111111111111111111111111")
	require.NoError(t, err)

	destAddr, err := lnchat.NewNodeFromString(
		"222222222222222222222222222222222222222222222222222222222222222222")
	require.NoError(t, err)

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: selfAddr.String(),
		},
	}

	paymentUpdateList := []lnchat.PaymentUpdate{
		{
			Payment: &lnchat.Payment{
				Hash:           "0000000000000000000000000000000000000000000000000000000000000001",
				Preimage:       "0000000000000000000000000000000000000000000000000000000000000010",
				Value:          lnchat.NewAmount(1200),
				CreationTimeNs: time.Now().Unix(),
				PaymentRequest: "",
				Status:         lnchat.PaymentSUCCEEDED,
				PaymentIndex:   1003,
				Htlcs: []lnchat.HTLCAttempt{
					{
						Route: lnchat.Route{
							TimeLock: 1000,
							Amt:      lnchat.NewAmount(1200),
							Hops: []lnchat.RouteHop{
								lnchat.RouteHop{
									ChannelID: 0x33,
									NodeID:    destAddr,
									Expiry:    100,
								},
							},
						},
						AttemptTimeNs: time.Now().Add(time.Second).Unix(),
						ResolveTimeNs: time.Now().Add(2 * time.Second).Unix(),
						Status:        lnrpc.HTLCAttempt_SUCCEEDED,
						Failure:       nil,
						Preimage:      []byte("00000000000000000000000000000001"),
					},
				},
			},
			Err: nil,
		},
		{
			Payment: &lnchat.Payment{
				Hash:           "0000000000000000000000000000000000000000000000000000000000000002",
				Preimage:       "0000000000000000000000000000000000000000000000000000000000000020",
				Value:          lnchat.NewAmount(1301),
				CreationTimeNs: time.Now().Unix(),
				PaymentRequest: "",
				Status:         lnchat.PaymentSUCCEEDED,
				PaymentIndex:   1004,
				Htlcs: []lnchat.HTLCAttempt{
					{
						Route: lnchat.Route{
							TimeLock: 1000,
							Amt:      lnchat.NewAmount(1301),
							Hops: []lnchat.RouteHop{
								lnchat.RouteHop{
									ChannelID: 0x33,
									NodeID:    destAddr,
									Expiry:    100,
								},
							},
						},
						AttemptTimeNs: time.Now().Add(time.Second).Unix(),
						ResolveTimeNs: time.Now().Add(2 * time.Second).Unix(),
						Status:        lnrpc.HTLCAttempt_SUCCEEDED,
						Failure:       nil,
						Preimage:      []byte("00000000000000000000000000000002"),
					},
				},
			},
			Err: nil,
		},
	}

	type paymentUpdateOp struct {
		data           lnchat.PaymentUpdate
		addPaymentsErr error
		payment        *model.Payment
	}

	cases := []struct {
		name                string
		subscrPayUpdatesErr error
		paymentUpdateOps    []paymentUpdateOp
	}{
		{
			name:                "success",
			subscrPayUpdatesErr: nil,
			paymentUpdateOps: []paymentUpdateOp{
				{
					data:           paymentUpdateList[0],
					addPaymentsErr: nil,
					payment: &model.Payment{
						PayerAddress: selfAddr.String(),
						PayeeAddress: destAddr.String(),
						Payment:      *paymentUpdateList[0].Payment,
					},
				},
				{
					data:           paymentUpdateList[1],
					addPaymentsErr: nil,
					payment: &model.Payment{
						PayerAddress: selfAddr.String(),
						PayeeAddress: destAddr.String(),
						Payment:      *paymentUpdateList[1].Payment,
					},
				},
			},
		},
		{
			name:                "subscription terminated",
			subscrPayUpdatesErr: nil,
			paymentUpdateOps:    []paymentUpdateOp{},
		},
		{
			name:                "AddPayment error",
			subscrPayUpdatesErr: nil,
			paymentUpdateOps: []paymentUpdateOp{
				{
					data:           paymentUpdateList[0],
					addPaymentsErr: fmt.Errorf("AddPayment error"),
					payment: &model.Payment{
						PayerAddress: selfAddr.String(),
						PayeeAddress: destAddr.String(),
						Payment:      *paymentUpdateList[0].Payment,
					},
				},
				{
					data:           paymentUpdateList[1],
					addPaymentsErr: nil,
					payment: &model.Payment{
						PayerAddress: selfAddr.String(),
						PayeeAddress: destAddr.String(),
						Payment:      *paymentUpdateList[1].Payment,
					},
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				testcaseTermination := make(chan bool)

				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(uint64(1), nil).Once()
				mockDB.On("GetLastPaymentIndex").Return(uint64(42), nil).Once()

				for _, op := range c.paymentUpdateOps {
					mockDB.On("AddPayments", op.payment).Return(
						op.addPaymentsErr).Once()
				}

				mockLNManager.On("SubscribeInvoiceUpdates", mock.Anything, uint64(1),
					mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

				paymentUpdateCh := func() <-chan lnchat.PaymentUpdate {
					ch := make(chan lnchat.PaymentUpdate)

					go func(channel chan lnchat.PaymentUpdate) {
						defer close(channel)

						for _, update := range c.paymentUpdateOps {
							channel <- update.data
						}

						<-testcaseTermination
					}(ch)
					return ch
				}()

				mockLNManager.On("SubscribePaymentUpdates", mock.Anything, uint64(42),
					mock.AnythingOfType("func(*lnchat.Payment) bool")).Return(
					paymentUpdateCh, c.subscrPayUpdatesErr).Once()

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

			payCh, err := app.SubscribePayments(ctxc)
			assert.NoError(t, err)

			// Verify that the expected updates are received from the channel
			expected := make([]*model.Payment, 0, len(c.paymentUpdateOps))
			var received []*model.Payment
			for _, u := range c.paymentUpdateOps {
				if u.payment != nil {
					expected = append(expected, u.payment)
				}
				if update, ok := <-payCh; ok {
					received = append(received, update)
				}
			}

			assert.ElementsMatch(t, expected, received)
		})
	}
}
