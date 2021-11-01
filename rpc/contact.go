package rpc

import (
	"context"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type contactServiceServer struct {
	Log *slog.Logger

	App *app.App
}

func (s *contactServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// AddContact adds a node as a contact to the database.
// The contact to be added is provided in the request,
// where its desired display name is also specified.
func (s *contactServiceServer) AddContact(ctx context.Context, req *pb.AddContactRequest) (*pb.AddContactResponse, error) {
	contact := contactInfoToContactModel(req.GetContact())

	savedContact, err := s.App.AddContact(ctx, &contact)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	responseContact := contactModelToContactInfo(*savedContact)

	return &pb.AddContactResponse{
		Contact: responseContact,
	}, nil
}

// GetContacts returns a contact list from database.
func (s *contactServiceServer) GetContacts(ctx context.Context, _ *pb.GetContactsRequest) (*pb.GetContactsResponse, error) {
	var contacts []model.Contact
	var err error

	// Search everything
	contacts, err = s.App.GetContacts(ctx)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	// Marshal data to result
	responseContacts := make([]*pb.ContactInfo, len(contacts))
	for i, c := range contacts {
		responseContacts[i] = contactModelToContactInfo(c)
	}

	return &pb.GetContactsResponse{
		Contacts: responseContacts,
	}, nil
}

// RemoveContactByID removes a contact from the database,
// based on the id request field.
func (s *contactServiceServer) RemoveContactByID(ctx context.Context, req *pb.RemoveContactByIDRequest) (*pb.RemoveContactResponse, error) {
	err := s.App.RemoveContactByID(ctx, req.GetId())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.RemoveContactResponse{}, nil
}

// RemoveContactByAddress removes a contact from the database,
// based on the address request field.
func (s *contactServiceServer) RemoveContactByAddress(ctx context.Context, req *pb.RemoveContactByAddressRequest) (*pb.RemoveContactResponse, error) {
	err := s.App.RemoveContactByAddress(ctx, req.GetAddress())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.RemoveContactResponse{}, nil
}

// NewContactServiceServer initializes a new contact service.
func NewContactServiceServer(app *app.App) pb.ContactServiceServer {
	return &contactServiceServer{
		Log: slog.NewLogger("contact-service"),
		App: app,
	}
}
