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
	indexIdxFunc := func(name string, value interface{}) ([]byte, error) {
		var p *Payment

		switch v := value.(type) {
		case *Payment:
			p = v
		// Workaround for badgerhold issue !43
		case **Payment:
			p = *v
		default:
			return nil, fmt.Errorf("PaymentIndex: expected Payment, got %T", value)
		}

		// Return the PaymentIndex.
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, p.PaymentIndex)

		return b, nil
	}

	return map[string]badgerhold.Index{
		"PaymentIndex": badgerhold.Index{
			IndexFunc: indexIdxFunc,
			Unique:    true,
		},
	}
}
