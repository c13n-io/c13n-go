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
