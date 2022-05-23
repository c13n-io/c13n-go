package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"

	"github.com/c13n-io/c13n-go/model"
)

var (
	// ErrDuplicateContact is returned when more than one contacts
	// were found, while at most one was expected.
	ErrDuplicateContact = fmt.Errorf("Duplicate contact")
	// ErrContactNotFound is returned in case a contact was not found.
	ErrContactNotFound = fmt.Errorf("Contact not found")
	// ErrContactAlreadyExists is returned in case the contact
	// already exists when attempting to insert it.
	ErrContactAlreadyExists = fmt.Errorf("Contact already exists")
)

// AddContact stores a contact.
func (db *bhDatabase) AddContact(contact *model.Contact) (
	c *model.Contact, err error) {

	uniq := badgerhold.Where("Node.Address").Eq(contact.Node.Address)

	if err = retryConflicts(db.bh.Badger().Update, func(txn *badger.Txn) error {
		switch _, err = db.findSingleContact(txn, uniq); err {
		case ErrContactNotFound:
		case nil:
			return ErrContactAlreadyExists
		default:
			return err
		}

		return db.bh.TxInsert(txn, badgerhold.NextSequence(), contact)
	}); err != nil {
		return nil, err
	}

	return contact, nil
}

// GetContact retrieves a contact.
func (db *bhDatabase) GetContact(address string) (contact *model.Contact, err error) {
	query := badgerhold.Where("Node.Address").Eq(address)

	err = db.bh.Badger().View(func(txn *badger.Txn) error {
		contact, err = db.findSingleContact(txn, query)
		return err
	})

	return
}

// GetContactById retrieves a contact by its key.
func (db *bhDatabase) GetContactByID(uid uint64) (contact *model.Contact, err error) {
	query := badgerhold.Where(badgerhold.Key).Eq(uid)

	err = db.bh.Badger().View(func(txn *badger.Txn) error {
		contact, err = db.findSingleContact(txn, query)
		return err
	})

	return
}

// RemoveContact removes a contact.
func (db *bhDatabase) RemoveContact(address string) (contact *model.Contact, err error) {
	addressQuery := badgerhold.Where("Node.Address").Eq(address)

	err = retryConflicts(db.bh.Badger().Update, func(txn *badger.Txn) error {
		contact, err = db.findSingleContact(txn, addressQuery)
		if err != nil {
			return err
		}

		return db.bh.TxDeleteMatching(txn, contact, addressQuery)
	})

	return
}

// RemoveContactByID removes a contact by its key.
func (db *bhDatabase) RemoveContactByID(uid uint64) (contact *model.Contact, err error) {
	query := badgerhold.Where(badgerhold.Key).Eq(uid)

	err = retryConflicts(db.bh.Badger().Update, func(txn *badger.Txn) error {
		contact, err = db.findSingleContact(txn, query)
		if err != nil {
			return err
		}

		return db.bh.TxDeleteMatching(txn, contact, query)
	})

	return
}

func (db *bhDatabase) findSingleContact(txn *badger.Txn,
	query *badgerhold.Query) (*model.Contact, error) {

	result := make([]model.Contact, 0)
	if err := db.bh.TxFind(txn, &result, query); err != nil {
		return nil, err
	}

	switch len(result) {
	case 1:
		return &result[0], nil
	case 0:
		return nil, ErrContactNotFound
	default:
		return nil, ErrDuplicateContact
	}
}

// GetContacts retrieves all contacts.
func (db *bhDatabase) GetContacts() ([]model.Contact, error) {
	contacts := make([]model.Contact, 0)

	if err := db.bh.Find(&contacts, nil); err != nil {
		return nil, err
	}

	return contacts, nil
}
