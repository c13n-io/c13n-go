package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/model"
)

func generateContact(nickname, alias, address string) model.Contact {
	return model.Contact{
		DisplayName: nickname,
		Node: model.Node{
			Alias:   alias,
			Address: address,
		},
	}
}

func TestAddContact(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contact := generateContact("alie", "alice",
		"012345678901234567890123456789012345678901234567890123456789012345")

	expected := contact
	expected.ID = 0

	res, err := db.AddContact(&contact)
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, res)
}

func TestAddContactDuplicateAddress(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contact := generateContact("alie", "alice",
		"012345678901234567890123456789012345678901234567890123456789012345")

	expected := contact
	expected.ID = 0

	res, err := db.AddContact(&contact)
	require.NoError(t, err)
	require.EqualValues(t, &expected, res)

	duplicate := generateContact("bobbie", "bob",
		"012345678901234567890123456789012345678901234567890123456789012345")

	duplResp, err := db.AddContact(&duplicate)
	assert.EqualError(t, err, ErrContactAlreadyExists.Error())
	assert.Nil(t, duplResp)
}

func TestAddContactIDs(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
		generateContact("carrie", "carol",
			"234567890123456789012345678901234567890123456789012345678901234567"),
	}

	for i, contact := range contacts {
		inserted, err := db.AddContact(&contact)
		assert.NoError(t, err)
		assert.EqualValues(t, &contact, inserted)
		assert.EqualValues(t, i, inserted.ID)
		assert.EqualValues(t, contact.ID, inserted.ID)
	}
}

func TestGetContact(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
		generateContact("carrie", "carol",
			"234567890123456789012345678901234567890123456789012345678901234567"),
	}

	for i := range contacts {
		inserted, err := db.AddContact(&contacts[i])
		require.NoError(t, err)
		require.EqualValues(t, &contacts[i], inserted)
	}

	bob := contacts[1]

	contact, err := db.GetContact(bob.Address)
	assert.NoError(t, err)
	assert.EqualValues(t, &bob, contact)

	carol := contacts[2]

	byID, err := db.GetContactByID(carol.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, &carol, byID)
}

func TestGetContactMissing(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
		generateContact("carrie", "carol",
			"234567890123456789012345678901234567890123456789012345678901234567"),
	}

	for i := range contacts {
		inserted, err := db.AddContact(&contacts[i])
		require.NoError(t, err)
		require.EqualValues(t, &contacts[i], inserted)
	}

	invalidAddr := "456789012345678901234567890123456789012345678901234567890123456789"
	invalidID := contacts[2].ID + 42

	missing, err := db.GetContact(invalidAddr)
	assert.EqualError(t, err, ErrContactNotFound.Error())
	assert.Nil(t, missing)

	byID, err := db.GetContactByID(invalidID)
	assert.EqualError(t, err, ErrContactNotFound.Error())
	assert.Nil(t, byID)
}

func TestRemoveContact(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
	}

	for i := range contacts {
		inserted, err := db.AddContact(&contacts[i])
		require.NoError(t, err)
		require.EqualValues(t, &contacts[i], inserted)
	}

	bob := contacts[1]

	deleted, err := db.RemoveContact(bob.Address)
	assert.NoError(t, err)
	assert.EqualValues(t, &bob, deleted)

	retrieved, err := db.GetContact(bob.Address)
	assert.EqualError(t, err, ErrContactNotFound.Error())
	assert.Nil(t, retrieved)
}

func TestRemoveContactById(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
	}

	for i := range contacts {
		inserted, err := db.AddContact(&contacts[i])
		require.NoError(t, err)
		require.EqualValues(t, &contacts[i], inserted)
	}

	bob := contacts[1]

	deleted, err := db.RemoveContactByID(bob.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, &bob, deleted)

	retrieved, err := db.GetContact(bob.Address)
	assert.EqualError(t, err, ErrContactNotFound.Error())
	assert.Nil(t, retrieved)
}

func TestGetContacts(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	contacts := []model.Contact{
		generateContact("alie", "alice",
			"012345678901234567890123456789012345678901234567890123456789012345"),
		generateContact("bobbie", "bob",
			"123456789012345678901234567890123456789012345678901234567890123456"),
		generateContact("carrie", "carol",
			"234567890123456789012345678901234567890123456789012345678901234567"),
	}

	for i := range contacts {
		inserted, err := db.AddContact(&contacts[i])
		require.NoError(t, err)
		require.EqualValues(t, &contacts[i], inserted)
	}

	list, err := db.GetContacts()
	assert.NoError(t, err)
	assert.EqualValues(t, contacts, list)
}
