package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

func generatePayment(t *testing.T, payee string, amt lnchat.Amount,
	status lnchat.PaymentStatus) *model.Payment {

	paymentIdx := lastPaymentIdx
	paymentCreationTime := lastPaymentCreationNs
	lastPaymentIdx++
	lastPaymentCreationNs += paymentCreationNsStep

	return &model.Payment{
		PayerAddress: selfAddress,
		PayeeAddress: payee,
		Payment: lnchat.Payment{
			Hash:           "fake preimage hash",
			Preimage:       generateHex(t, 32),
			PaymentRequest: "fake payment request",
			Value:          amt,
			Status:         status,
			CreationTimeNs: paymentCreationTime,
			PaymentIndex:   paymentIdx,
		},
	}
}

func TestAddPayments(t *testing.T) {
	dest := "444444444444444444444444444444444444444444444444444444444444444444"
	testPayments := []*model.Payment{
		generatePayment(t, dest, 5432, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dest, 4321, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dest, 3210, lnchat.PaymentSUCCEEDED),
	}

	cases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "success",
			test: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				err := db.AddPayments(testPayments[0])
				assert.NoError(t, err)
			},
		},
		{
			name: "multiple",
			test: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				err := db.AddPayments(testPayments...)
				assert.NoError(t, err)

				payments, err := db.GetPayments(model.PageOptions{})
				assert.NoError(t, err)
				assert.Len(t, payments, len(testPayments))
			},
		},
		{
			name: "duplicate",
			test: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				err := db.AddPayments(testPayments[0])
				require.NoError(t, err)

				expectedErr := alreadyExists(testPayments[0])

				err = db.AddPayments(testPayments[0])
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.test)
	}
}

func TestGetPayments(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	dests := []string{
		"111111111111111111111111111111111111111111111111111111111111111111",
		"222222222222222222222222222222222222222222222222222222222222222222",
		"333333333333333333333333333333333333333333333333333333333333333333",
	}
	payments := []*model.Payment{
		generatePayment(t, dests[1], 1234, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[0], 2345, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[2], 3456, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[1], 4567, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[1], 5678, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[0], 6789, lnchat.PaymentSUCCEEDED),
		generatePayment(t, dests[1], 7890, lnchat.PaymentSUCCEEDED),
	}

	for _, pmnt := range payments {
		err := db.AddPayments(pmnt)
		require.NoError(t, err)
	}

	reversePayments := func(pmnts []*model.Payment) []*model.Payment {
		length := len(pmnts)
		res := make([]*model.Payment, length)
		for i, el := range pmnts {
			res[length-i-1] = el
		}

		return res
	}(payments)

	cases := []struct {
		name             string
		pageOpts         model.PageOptions
		expectedPayments []*model.Payment
		expectedErr      error
	}{
		{
			name:             "all",
			pageOpts:         model.PageOptions{},
			expectedPayments: payments[:],
		},
		{
			name: "with start",
			pageOpts: model.PageOptions{
				LastID: payments[1].PaymentIndex,
			},
			expectedPayments: payments[1:],
		},
		{
			name: "reverse with start",
			pageOpts: model.PageOptions{
				LastID:  payments[5].PaymentIndex,
				Reverse: true,
			},
			expectedPayments: reversePayments[6-5:],
		},
		{
			name: "with size",
			pageOpts: model.PageOptions{
				PageSize: 3,
			},
			expectedPayments: payments[:3],
		},
		{
			name: "reverse with size",
			pageOpts: model.PageOptions{
				PageSize: 4,
				Reverse:  true,
			},
			expectedPayments: reversePayments[:4],
		},
		{
			name: "with start and size",
			pageOpts: model.PageOptions{
				LastID:   payments[1].PaymentIndex,
				PageSize: 4,
			},
			expectedPayments: payments[1:5],
		},
		{
			name: "reverse with start and size",
			pageOpts: model.PageOptions{
				LastID:   payments[5].PaymentIndex,
				PageSize: 3,
				Reverse:  true,
			},
			expectedPayments: reversePayments[6-5 : 6-5+3],
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pmnts, err := db.GetPayments(c.pageOpts)
			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.Equal(t, c.expectedPayments, pmnts)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, pmnts)
			}
		})
	}
}
