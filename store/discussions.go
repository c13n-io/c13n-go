package store

import (
	"fmt"
	"sort"

	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/model"
)

var (
	// ErrDuplicateDiscussion is returned when more then one discussions
	// were found, while at most one was expected.
	ErrDuplicateDiscussion = fmt.Errorf("Duplicate discussion")
	// ErrDiscussionNotFound is returned in case a discussion was not found.
	ErrDiscussionNotFound = fmt.Errorf("Discussion not found")
	// ErrDiscussionAlreadyExists is returned in case the discussion
	// already exists when attempting to insert it.
	ErrDiscussionAlreadyExists = fmt.Errorf("Discussion already exists")
	// ErrMessageNotFound is returned in case a message id was not found.
	ErrMessageNotFound = fmt.Errorf("Message not found")
	// ErrMessageInvalidDisc is returned in case a message does
	// not belong to a discussion.
	ErrMessageInvalidDisc = fmt.Errorf("Message does not belong to discussion")
)

// AddDiscussion stores a discussion.
func (db *bhDatabase) AddDiscussion(discussion *model.Discussion) (*model.Discussion, error) {
	// Sort participant slice for querying by participants.
	sort.Strings(discussion.Participants)

	err := db.bh.Insert(badgerhold.NextSequence(), discussion)
	if err == badgerhold.ErrUniqueExists {
		return nil, ErrDiscussionAlreadyExists
	}

	return discussion, err
}

// GetDiscussion retrieves a discussion.
func (db *bhDatabase) GetDiscussion(uid uint64) (discussion *model.Discussion, err error) {
	query := badgerhold.Where(badgerhold.Key).Eq(uid)

	err = db.bh.Badger().View(func(txn *badger.Txn) error {
		discussion, err = db.findSingleDiscussion(txn, query)
		return err
	})

	return
}

// GetDiscussionByParticipants retrieves a discussion based on its participant set.
func (db *bhDatabase) GetDiscussionByParticipants(
	participants []string) (discussion *model.Discussion, err error) {

	sort.Strings(participants)
	query := badgerhold.Where("Participants").Eq(participants)

	err = db.bh.Badger().View(func(txn *badger.Txn) error {
		discussion, err = db.findSingleDiscussion(txn, query)
		return err
	})

	return
}

// RemoveDiscussion removes a discussion.
func (db *bhDatabase) RemoveDiscussion(uid uint64) (discussion *model.Discussion, err error) {
	query := badgerhold.Where(badgerhold.Key).Eq(uid)

	err = db.bh.Badger().Update(func(txn *badger.Txn) error {
		discussion, err = db.findSingleDiscussion(txn, query)
		if err != nil {
			return err
		}

		return db.bh.TxDelete(txn, uid, discussion)
	})

	return
}

// UpdateDiscussionLastRead updates a discussion's last read message
// with the provided messsage id, if the message id belongs to the discussion.
func (db *bhDatabase) UpdateDiscussionLastRead(uid uint64, readMsgID uint64) error {
	query := badgerhold.Where(badgerhold.Key).Eq(uid)

	err := db.bh.Badger().Update(func(txn *badger.Txn) error {
		// Verify that the message belongs to the discussion.
		msg := &model.RawMessage{}
		if err := db.bh.TxGet(txn, readMsgID, msg); err != nil {
			return ErrMessageNotFound
		}
		if msg.DiscussionID != uid {
			return ErrMessageInvalidDisc
		}

		// Update the stored discussion.
		err := db.bh.TxUpdateMatching(txn, &model.Discussion{}, query, func(record interface{}) error {
			disc, ok := record.(*model.Discussion)
			if !ok {
				return ErrDiscussionNotFound
			}

			disc.LastReadID = readMsgID

			return nil
		})
		if err != nil {
			return ErrDiscussionNotFound
		}

		return nil
	})

	return err
}

func (db *bhDatabase) findSingleDiscussion(txn *badger.Txn,
	query *badgerhold.Query) (*model.Discussion, error) {

	result := make([]model.Discussion, 0)
	if err := db.bh.TxFind(txn, &result, query); err != nil {
		return nil, err
	}

	switch len(result) {
	case 1:
		return &result[0], nil
	case 0:
		return nil, ErrDiscussionNotFound
	default:
		return nil, ErrDuplicateDiscussion
	}
}

// GetDiscussions retrieves discussions, respecting pagination.
// The seekIndex and pageSize parameters control
// the start and length of the requested range.
// seekIndex of 0 corresponds to starting from the first discussion, while
// pageSize of 0 corresponds to no length limit for the result.
func (db *bhDatabase) GetDiscussions(seekIndex, pageSize uint64) ([]model.Discussion, error) {
	query := (&badgerhold.Query{}).Skip(int(seekIndex)).Limit(int(pageSize))
	discussions := make([]model.Discussion, 0)

	if err := db.bh.Find(&discussions, query); err != nil {
		return nil, err
	}

	return discussions, nil
}
