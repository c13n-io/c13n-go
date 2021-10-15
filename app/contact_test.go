package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/c13n-io/c13n-backend/lnchat"
	lnmock "github.com/c13n-io/c13n-backend/lnchat/mocks"
	"github.com/c13n-io/c13n-backend/model"
	"github.com/c13n-io/c13n-backend/store"
	dbmock "github.com/c13n-io/c13n-backend/store/mocks"
)

func TestAddContact(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	cases := []struct {
		name            string
		contactToAdd    *model.Contact
		addContactResp  *model.Contact
		addContactErr   error
		expectedContact *model.Contact
		expectedErr     error
	}{
		{
			name: "Add contact",
			contactToAdd: &model.Contact{
				DisplayName: "contact nickname",
				Node: model.Node{
					Alias:   "network alias",
					Address: "121212121212121212121212121212121212121212121212121212121212121212",
				},
			},
			addContactResp: &model.Contact{
				ID:          12,
				DisplayName: "contact nickname",
				Node: model.Node{
					Alias:   "network alias",
					Address: "121212121212121212121212121212121212121212121212121212121212121212",
				},
			},
			addContactErr: nil,
			expectedContact: &model.Contact{
				ID:          12,
				DisplayName: "contact nickname",
				Node: model.Node{
					Alias:   "network alias",
					Address: "121212121212121212121212121212121212121212121212121212121212121212",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Add existing contact",
			contactToAdd: &model.Contact{
				DisplayName: "contact nickname",
				Node: model.Node{
					Alias:   "network alias",
					Address: "121212121212121212121212121212121212121212121212121212121212121212",
				},
			},
			addContactResp:  nil,
			addContactErr:   store.ErrContactAlreadyExists,
			expectedContact: nil,
			expectedErr: Error{
				Kind:    ContactAlreadyExists,
				Err:     store.ErrContactAlreadyExists,
				details: "AddContact",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(
					uint64(1), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates",
					mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockDB.On("AddContact", c.contactToAdd).Return(
					c.addContactResp, c.addContactErr).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			contact, err := app.AddContact(ctxt, c.contactToAdd)

			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
				assert.EqualValues(t, c.expectedContact, contact)
			default:
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, contact)
			}
		})
	}
}

func TestGetContactsSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	contactList := []model.Contact{
		{
			ID:          0,
			DisplayName: "nickname A",
			Node: model.Node{
				Alias:   "alice",
				Address: "111111111111111111111111111111111111111111111111111111111111111111",
			},
		},
		{
			ID:          1,
			DisplayName: "nickname B",
			Node: model.Node{
				Alias:   "bob",
				Address: "222222222222222222222222222222222222222222222222222222222222222222",
			},
		},
		{
			ID:          2,
			DisplayName: "nickname C",
			Node: model.Node{
				Alias:   "carol",
				Address: "333333333333333333333333333333333333333333333333333333333333333333",
			},
		},
		{
			ID:          3,
			DisplayName: "nickname D",
			Node: model.Node{
				Alias:   "goliath",
				Address: "543543543543543543543543543543543543543543543543543543543543543543",
			},
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockDB.On("GetContacts").Return(contactList, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	contacts, err := app.GetContacts(ctxt)
	assert.NoError(t, err)
	assert.EqualValues(t, contactList, contacts)
}

func TestRemoveContactByAddress(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	cases := []struct {
		name              string
		contactAddr       string
		removeContactResp *model.Contact
		removeContactErr  error
		expectedErr       error
	}{
		{
			name:        "Success",
			contactAddr: "111111111111111111111111111111111111111111111111111111111111111111",
			removeContactResp: &model.Contact{
				ID:          32,
				DisplayName: "nickname",
				Node: model.Node{
					Alias:   "bob",
					Address: "111111111111111111111111111111111111111111111111111111111111111111",
				},
			},
			removeContactErr: nil,
			expectedErr:      nil,
		},
		{
			name:              "Not found",
			contactAddr:       "121212121212121212121212121212121212121212121212121212121212121212",
			removeContactResp: nil,
			removeContactErr:  store.ErrContactNotFound,
			expectedErr: Error{
				Kind:    ContactNotFound,
				Err:     store.ErrContactNotFound,
				details: "RemoveContact",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
				*lnmock.LightManager, *dbmock.Database, func()) {

				// Mock self info
				mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

				mockDB.On("GetLastInvoiceIndex").Return(
					uint64(1), nil).Once()

				mockLNManager.On("SubscribeInvoiceUpdates",
					mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

				mockDB.On("Close").Return(nil).Once()
				mockLNManager.On("Close").Return(nil).Once()

				mockDB.On("RemoveContact", c.contactAddr).Return(
					c.removeContactResp, c.removeContactErr).Once()

				mockStopFunc := func() {}

				return mockLNManager, mockDB, mockStopFunc
			}

			app, appTestStartFunc, appTestStopFunc :=
				createInitializedApp(t, mockInstaller)

			appTestStartFunc()
			defer appTestStopFunc()

			ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()

			err := app.RemoveContactByAddress(ctxt, c.contactAddr)

			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}
