package store

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/model"
)

// AddPayments stores a list of payments.
func (db *bhDatabase) AddPayments(payments ...*model.Payment) error {
	if len(payments) <= 0 {
		return nil
	}

	return retryConflicts(db.bh.Badger().Update, func(txn *badger.Txn) error {
		for _, payment := range payments {
			paymentKey := payment.PaymentIndex

			switch err := db.bh.TxInsert(txn, paymentKey, payment); err {
			case badgerhold.ErrKeyExists:
				return alreadyExists(payment)
			default:
				return err
			}
		}
		return nil
	})
}

// GetLastPaymentIndex retrieves the last payment index present in the database.
func (db *bhDatabase) GetLastPaymentIndex() (paymentIdx uint64, err error) {
	p := new(model.Payment)

	switch result, err := db.bh.FindAggregate(p, nil); err {
	case nil:
		if result[0].Count() > 0 {
			result[0].Max("PaymentIndex", p)
			return p.PaymentIndex, nil
		}
	default:
		return 0, err
	}

	return
}
