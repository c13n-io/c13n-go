package app

import (
	"context"

	"github.com/c13n-io/c13n-backend/model"
)

// AddContact adds a contact to the database if it doesn't exist.
func (app *App) AddContact(_ context.Context, contact *model.Contact) (*model.Contact, error) {
	contact, err := app.Database.AddContact(contact)
	if err != nil {
		return nil, newErrorf(err, "AddContact")
	}

	return contact, nil
}

// GetContacts returns all contacts stored in database.
func (app *App) GetContacts(_ context.Context) ([]model.Contact, error) {
	// Fetch contacts from database
	contacts, err := app.Database.GetContacts()
	if err != nil {
		return nil, newErrorf(err, "GetContacts")
	}

	return contacts, nil
}

// RemoveContactByID removes the contact matching the passed id from database.
func (app *App) RemoveContactByID(_ context.Context, id uint64) error {
	_, err := app.Database.RemoveContactByID(id)

	return newErrorf(err, "RemoveContactByID")
}

// RemoveContactByAddress removes the contact matching the passed address from database.
func (app *App) RemoveContactByAddress(_ context.Context, address string) error {
	_, err := app.Database.RemoveContact(address)

	return newErrorf(err, "RemoveContact")
}
