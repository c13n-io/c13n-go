package store

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/model"
)

// AddInvoice stores an invoice.
func (db *bhDatabase) AddInvoice(inv *model.Invoice) error {
	return retryConflicts(db.bh.Badger().Update, func(txn *badger.Txn) error {
		invoiceKey := inv.SettleIndex
		switch err := db.bh.TxInsert(txn, invoiceKey, inv); err {
		case badgerhold.ErrKeyExists:
			return alreadyExists(inv)
		default:
			return err
		}
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

// GetInvoices retrieves invoices, based on the provided pagination options.
func (db *bhDatabase) GetInvoices(pageOpts model.PageOptions) ([]*model.Invoice, error) {
	q := &badgerhold.Query{}
	switch {
	case pageOpts.LastID != 0 && pageOpts.Reverse:
		q = badgerhold.Where("SettleIndex").Le(pageOpts.LastID)
	case pageOpts.LastID != 0:
		q = badgerhold.Where("SettleIndex").Ge(pageOpts.LastID)
	}
	q = q.SortBy("SettleIndex").Limit(int(pageOpts.PageSize))
	if pageOpts.Reverse {
		q = q.Reverse()
	}

	var invoices []*model.Invoice
	if err := db.bh.Find(&invoices, q); err != nil {
		return nil, err
	}

	return invoices, nil
}
