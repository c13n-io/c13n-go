package model

import (
	"encoding/binary"
	"fmt"

	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/lnchat"
)

// Invoice embeds lnchat.Invoice,
// implementing the badgerhold.Storer interface.
type Invoice struct {
	// Since only the invoice creator has access to the Invoice,
	// the CreatorAddress is the Lightning address of the underlying node.
	CreatorAddress string
	// The embedded invoice.
	lnchat.Invoice
}

// Type satisfies badgerhold.Storer interface for Invoice type.
func (i *Invoice) Type() string {
	return "Invoice"
}

// Indexes satisfies badgerhold.Storer interface for Invoice type.
func (i *Invoice) Indexes() map[string]badgerhold.Index {
	getInvoice := func(value interface{}) (inv *Invoice, ok bool) {
		inv, ok = value.(*Invoice)
		return
	}

	return map[string]badgerhold.Index{
		"SettleIndex": badgerhold.Index{
			IndexFunc: func(_ string, value interface{}) ([]byte, error) {
				inv, ok := getInvoice(value)
				if !ok {
					return nil, fmt.Errorf("InvoiceSettleIndex:"+
						" expected Invoice, got %T", value)
				}

				b := make([]byte, 8)
				binary.BigEndian.PutUint64(b, inv.SettleIndex)
				return b, nil
			},
			Unique: true,
		},
		"PreimageIndex": badgerhold.Index{
			IndexFunc: func(_ string, value interface{}) ([]byte, error) {
				inv, ok := getInvoice(value)
				if !ok {
					return nil, fmt.Errorf("InvoicePreimageIndex:"+
						" expected Invoice, got %T", value)
				}

				b := make([]byte, len(inv.Preimage))
				copy(b, inv.Preimage)
				return b, nil
			},
			Unique: true,
		},
	}
}
