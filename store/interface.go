package store

//go:generate mockery --dir=. --output=./mocks --outpkg=storemock --name=Database

import (
	"github.com/c13n-io/c13n-go/model"
)

// Database provices the generic interface for database operations.
type Database interface {
	// Contacts
	AddContact(c *model.Contact) (contact *model.Contact, err error)
	GetContact(address string) (*model.Contact, error)
	GetContactByID(uid uint64) (*model.Contact, error)
	RemoveContact(address string) (*model.Contact, error)
	RemoveContactByID(uid uint64) (*model.Contact, error)
	GetContacts() ([]model.Contact, error)

	// Discussions
	AddDiscussion(disc *model.Discussion) (discussion *model.Discussion, err error)
	GetDiscussion(uid uint64) (*model.Discussion, error)
	GetDiscussionByParticipants(participants []string) (*model.Discussion, error)
	RemoveDiscussion(uid uint64) (*model.Discussion, error)
	GetDiscussions(seekIndex, pageSize uint64) ([]model.Discussion, error)
	UpdateDiscussionLastRead(uid uint64, readMsgID uint64) error

	// Invoices-Payments
	AddInvoice(inv *model.Invoice) error
	AddPayments(payments ...*model.Payment) error
	GetLastInvoiceIndex() (invSettleIndex uint64, err error)
	GetLastPaymentIndex() (paymentIndex uint64, err error)
	GetInvoices(pageOpts model.PageOptions) ([]*model.Invoice, error)
	GetPayments(pageOpts model.PageOptions) ([]*model.Payment, error)
	AddRawMessage(*model.RawMessage) error
	GetMessages(discussionUID uint64,
		pageOpts model.PageOptions) ([]model.MessageAggregate, error)

	// Close closes the database
	Close() error
}
