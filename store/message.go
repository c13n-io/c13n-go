package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/model"
)

// AddRawMessage stores a raw message under a discussion
// and updates the last discussion message.
// An error is returned if its associated invoice or payment indexes are missing.
func (db *bhDatabase) AddRawMessage(rawMsg *model.RawMessage) error {
	return db.bh.Badger().Update(func(txn *badger.Txn) error {
		// Verify the existence of the associated invoice or payment
		invIdx := rawMsg.InvoiceSettleIndex
		paymentIdxs := rawMsg.PaymentIndexes
		switch {
		case len(paymentIdxs) == 0 && invIdx == 0:
			return fmt.Errorf("message not associated with invoice or payment")
		case invIdx != 0:
			if _, err := db.findInvoice(txn, invIdx); err != nil {
				return fmt.Errorf("could not retrieve associated invoice: %w", err)
			}
		case len(paymentIdxs) != 0:
			if _, err := db.findPayments(txn, paymentIdxs...); err != nil {
				return fmt.Errorf("could not retrieve associated payments: %w", err)
			}
		}

		// Verify the existence of the associated discussion
		discQuery := badgerhold.Where(badgerhold.Key).Eq(rawMsg.DiscussionID)
		if _, err := db.findSingleDiscussion(txn, discQuery); err != nil {
			return fmt.Errorf("could not retrieve associated discussion: %w", err)
		}

		rawMsg.WithTimestamp(getCurrentTime())

		// Insert the raw message
		if err := db.bh.TxInsert(txn, badgerhold.NextSequence(), rawMsg); err != nil {
			return err
		}

		// Update the discussion last message id
		return db.bh.TxUpdateMatching(txn, &model.Discussion{}, discQuery,
			func(record interface{}) error {
				disc, ok := record.(*model.Discussion)
				if !ok {
					return ErrDiscussionNotFound
				}

				disc.LastMessageID = rawMsg.ID
				return nil
			})
	})
}

func (db *bhDatabase) findInvoice(txn *badger.Txn,
	invoiceIdx uint64) (*model.Invoice, error) {

	invQuery := badgerhold.Where(badgerhold.Key).Eq(invoiceIdx)

	invs := make([]model.Invoice, 0)
	if err := db.bh.TxFind(txn, &invs, invQuery); err != nil {
		return nil, err
	}

	switch len(invs) {
	case 1:
		return &invs[0], nil
	case 0:
		return nil, fmt.Errorf("invoice not found")
	default:
		return nil, fmt.Errorf("duplicate invoice found")
	}
}

func (db *bhDatabase) findPayments(txn *badger.Txn,
	paymentIdxs ...uint64) ([]model.Payment, error) {

	// Slice type conversion
	payIdxs := make([]interface{}, len(paymentIdxs))
	for i := range paymentIdxs {
		payIdxs[i] = paymentIdxs[i]
	}

	// NOTE: Use of membership criterion (.In) is not playing nice
	// with either the use of (.Index) or badgerhold.Key
	payQuery := badgerhold.Where("PaymentIndex").In(payIdxs...)

	pays := make([]model.Payment, 0)
	if err := db.bh.TxFind(txn, &pays, payQuery); err != nil {
		return nil, fmt.Errorf("could not retrieve payment: %w", err)
	}

	resultIdxs := make([]uint64, len(pays))
	for i, pay := range pays {
		resultIdxs[i] = pay.PaymentIndex
	}
	switch sameUnorderedIDSlice(paymentIdxs, resultIdxs) {
	case true:
		return pays, nil
	default:
		return nil, fmt.Errorf("missing or mismatched payment detected")
	}
}

func sameUnorderedIDSlice(x, y []uint64) bool {
	if len(x) != len(y) {
		return false
	}

	// diff is a frequency map.
	diff := make(map[uint64]int, len(x))
	// Populate diff map with x's elements.
	for _, xx := range x {
		diff[xx]++
	}
	// Iterate on y, removing occurences as they are seen.
	for _, yy := range y {
		if _, ok := diff[yy]; !ok {
			return false
		}
		diff[yy]--
		// Delete to avoid another loop for the return condition.
		if diff[yy] == 0 {
			delete(diff, yy)
		}
	}

	return len(diff) == 0
}
