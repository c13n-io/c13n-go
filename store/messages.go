package store

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/timshannon/badgerhold"

	"github.com/c13n-io/c13n-backend/model"
)

// MessageAggregate represents a raw discussion message
// along with the invoice or payments it references.
type MessageAggregate struct {
	// The raw message.
	RawMessage *model.RawMessage
	// The associated invoice (if any; only valid for incoming messages).
	Invoice *model.Invoice
	// The associated payments (if any; only valid for outgoind messages).
	Payments []*model.Payment
}

func newMsgAggregate(raw model.RawMessage,
	inv *model.Invoice, pays []model.Payment) MessageAggregate {

	var payments []*model.Payment

	if len(pays) > 0 {
		payments = make([]*model.Payment, len(pays))
	}
	for i, pay := range pays {
		payments[i] = &pay
	}

	return MessageAggregate{
		RawMessage: &raw,
		Invoice:    inv,
		Payments:   payments,
	}
}

// GetMessages retrieves messages belonging to a discussion.
// The pageOpts parameter controls the requested message range.
func (db *bhDatabase) GetMessages(discussionUID uint64,
	pageOpts model.PageOptions) ([]MessageAggregate, error) {

	if pageOpts.Reverse && pageOpts.LastID == 0 {
		return nil, fmt.Errorf("reverse pagination without anchor is disallowed")
	}

	var messages []MessageAggregate
	if err := db.bh.Badger().View(func(txn *badger.Txn) error {
		discQuery := badgerhold.Where(badgerhold.Key).Eq(discussionUID)
		if _, err := db.findSingleDiscussion(txn, discQuery); err != nil {
			return err
		}

		// RawMessage retrieval query
		query := badgerhold.Where("DiscussionID").Eq(discussionUID).
			Index("DiscussionID")
		switch {
		case pageOpts.Reverse && pageOpts.LastID > 0:
			query = query.And(badgerhold.Key).Le(pageOpts.LastID).
				SortBy("Timestamp").Reverse()
		case pageOpts.LastID > 0:
			query = query.And(badgerhold.Key).Ge(pageOpts.LastID)
		}
		query = query.Limit(int(pageOpts.PageSize))

		// Retrieve the raw messages
		raws := make([]model.RawMessage, 0)
		if err := db.bh.TxFind(txn, &raws, query); err != nil {
			return err
		}
		messages = make([]MessageAggregate, len(raws))

		for i, raw := range raws {
			switch {
			case raw.InvoiceSettleIndex != 0:
				inv, err := db.findInvoice(txn, raw.InvoiceSettleIndex)
				if err != nil {
					return fmt.Errorf("could not retrieve invoice "+
						"associated to message %d: %w", raw.ID, err)
				}

				messages[i] = newMsgAggregate(raws[i], inv, nil)
			case raw.PaymentIndexes != nil:
				pays, err := db.findPayments(txn, raw.PaymentIndexes...)
				if err != nil {
					return fmt.Errorf("could not retrieve payments "+
						"associated with message %d: %w", raw.ID, err)
				}

				messages[i] = newMsgAggregate(raws[i], nil, pays)
			default:
				return fmt.Errorf("stored message not " +
					"associated with invoice or payments")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if pageOpts.Reverse {
		reverseMsgs := make([]MessageAggregate, len(messages))
		for i := len(messages) - 1; i >= 0; i-- {
			reverseMsgs[len(messages)-1-i] = messages[i]
		}

		return reverseMsgs, nil
	}

	return messages, nil
}
