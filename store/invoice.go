package store

import (
	"github.com/dgraph-io/badger/v3"

	"github.com/c13n-io/c13n-go/model"
)

// AddInvoice stores an invoice.
func (db *bhDatabase) AddInvoice(inv *model.Invoice) error {
	return db.bh.Badger().Update(func(txn *badger.Txn) error {
		invoiceKey := inv.SettleIndex
		return db.bh.TxInsert(txn, invoiceKey, inv)
	})
}

// GetLastInvoiceIndex retrieves the last invoice index present in the database.
func (db *bhDatabase) GetLastInvoiceIndex() (invoiceSettleIdx uint64, err error) {
	inv := new(model.Invoice)

	switch result, err := db.bh.FindAggregate(inv, nil); err {
	case nil:
		if result[0].Count() > 0 {
			result[0].Max("SettleIndex", inv)
			return inv.SettleIndex, nil
		}
	default:
		return 0, err
	}

	return
}
