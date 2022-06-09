package model

import (
	"encoding/binary"
	"fmt"

	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/lnchat"
)

// Payment embeds lnchat.Payment,
// implementing the badgerhold.Storer interface.
type Payment struct {
	// The Lightningaddress of the payer.
	PayerAddress string
	// The Lightning address of the payee.
	PayeeAddress string
	// The embedded payment.
	lnchat.Payment
}

// Type satisfies badgerhold.Storer interface for Payment type.
func (p *Payment) Type() string {
	return "Payment"
}

// Indexes satisfies badgerhold.Storer interface for Payment type.
func (p *Payment) Indexes() map[string]badgerhold.Index {
	getPayment := func(value interface{}) (pmnt *Payment, ok bool) {
		pmnt, ok = value.(*Payment)
		return
	}

	return map[string]badgerhold.Index{
		"PaymentIndex": badgerhold.Index{
			IndexFunc: func(_ string, value interface{}) ([]byte, error) {
				pmnt, ok := getPayment(value)
				if !ok {
					return nil, fmt.Errorf("PaymentIndex: "+
						"expected Payment, got %T", value)
				}

				b := make([]byte, 8)
				binary.BigEndian.PutUint64(b, pmnt.PaymentIndex)
				return b, nil
			},
			Unique: true,
		},
	}
}
