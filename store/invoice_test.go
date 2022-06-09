package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

func generateInvoice(t *testing.T, amt lnchat.Amount,
	state lnchat.InvoiceState) *model.Invoice {

	var invAddIdx, invSettleIdx uint64
	var invCreateTime, invSettleTime int64
	invAddIdx = lastInvoiceAddIdx
	lastInvoiceAddIdx++
	invCreateTime = lastInvoiceCreatedSec
	lastInvoiceCreatedSec += invoiceSettleSecStep
	if state == lnchat.InvoiceSETTLED {
		invSettleIdx = lastInvoiceSettleIdx
		lastInvoiceSettleIdx++
		invSettleTime = invCreateTime + invoiceSettleSecDiff
	}

	return &model.Invoice{
		CreatorAddress: selfAddress,
		Invoice: lnchat.Invoice{
			Hash:           "fake preimage hash",
			Preimage:       generateBytes(t, 32),
			PaymentRequest: "fake payment request",
			Value:          amt,
			AmtPaid:        amt,
			CreatedTimeSec: invCreateTime,
			SettleTimeSec:  invSettleTime,
			State:          state,
			AddIndex:       invAddIdx,
			SettleIndex:    invSettleIdx,
		},
	}
}

func TestAddInvoice(t *testing.T) {
	testInvoice := generateInvoice(t, 5432, lnchat.InvoiceSETTLED)

	cases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "success",
			test: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				err := db.AddInvoice(testInvoice)
				assert.NoError(t, err)
			},
		},
		{
			name: "duplicate",
			test: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				err := db.AddInvoice(testInvoice)
				require.NoError(t, err)

				expectedErr := alreadyExists(testInvoice)

				err = db.AddInvoice(testInvoice)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.test)
	}
}

func TestGetInvoices(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	invoices := []*model.Invoice{
		generateInvoice(t, 1234, lnchat.InvoiceSETTLED),
		generateInvoice(t, 2345, lnchat.InvoiceSETTLED),
		generateInvoice(t, 3456, lnchat.InvoiceSETTLED),
		generateInvoice(t, 4567, lnchat.InvoiceSETTLED),
		generateInvoice(t, 5678, lnchat.InvoiceSETTLED),
		generateInvoice(t, 6789, lnchat.InvoiceSETTLED),
		generateInvoice(t, 7890, lnchat.InvoiceSETTLED),
	}

	for _, invoice := range invoices {
		err := db.AddInvoice(invoice)
		require.NoError(t, err)
	}

	reverseInvoices := func(invs []*model.Invoice) []*model.Invoice {
		length := len(invs)
		res := make([]*model.Invoice, length)
		for i, el := range invs {
			res[length-i-1] = el
		}

		return res
	}(invoices)

	cases := []struct {
		name             string
		pageOpts         model.PageOptions
		expectedInvoices []*model.Invoice
		expectedErr      error
	}{
		{
			name:             "all",
			pageOpts:         model.PageOptions{},
			expectedInvoices: invoices[:],
		},
		{
			name: "with start",
			pageOpts: model.PageOptions{
				LastID: invoices[1].SettleIndex,
			},
			expectedInvoices: invoices[1:],
		},
		{
			name: "reverse with start",
			pageOpts: model.PageOptions{
				LastID:  invoices[5].SettleIndex,
				Reverse: true,
			},
			expectedInvoices: reverseInvoices[6-5:],
		},
		{
			name: "with size",
			pageOpts: model.PageOptions{
				PageSize: 3,
			},
			expectedInvoices: invoices[:3],
		},
		{
			name: "reverse with size",
			pageOpts: model.PageOptions{
				PageSize: 4,
				Reverse:  true,
			},
			expectedInvoices: reverseInvoices[:4],
		},
		{
			name: "with start and size",
			pageOpts: model.PageOptions{
				LastID:   invoices[1].SettleIndex,
				PageSize: 3,
			},
			expectedInvoices: invoices[1:4],
		},
		{
			name: "reverse with start and size",
			pageOpts: model.PageOptions{
				LastID:   invoices[5].SettleIndex,
				PageSize: 3,
				Reverse:  true,
			},
			expectedInvoices: reverseInvoices[6-5 : 6-5+3],
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			invs, err := db.GetInvoices(c.pageOpts)
			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.Equal(t, c.expectedInvoices, invs)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, invs)
			}
		})
	}
}
