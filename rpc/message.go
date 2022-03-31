package rpc

import (
	"context"
	"fmt"

	"github.com/c13n-io/c13n-go/app"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type messageServiceServer struct {
	Log *slog.Logger

	App *app.App

	pb.UnimplementedMessageServiceServer
}

func (s *messageServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// EstimateMessage creates a message containing the specified payload to be sent
// to the specified receiver.
// The returned message contains the calculated route, the fees associated with that route,
// as well as the probability of arrival to the receiver.
func (s *messageServiceServer) EstimateMessage(ctx context.Context, req *pb.EstimateMessageRequest) (*pb.EstimateMessageResponse, error) {
	msg, err := estimateMessageRequestToMessageModel(req)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}
	opts := messageOptionsFromRequest(req.GetOptions())

	estimation, err := s.App.EstimatePayment(ctx,
		msg.Payload, msg.AmtMsat, msg.DiscussionID, opts)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	resp, err := messageModelToEstimateMessageResponse(estimation)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return resp, nil
}

// SendMessage sends the provided message.
// If the message contains a route it is used, otherwise the message is sent
// respecting the established payment and message options.
// If the message contains a payment request, the payment is not spontaneous
// but instead the invoice is paid. In that case, the message is associated
// with the recipient's discussion, which is created if it doesn't exist.
func (s *messageServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	msg, err := sendMessageRequestToMessageModel(req)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}
	opts := messageOptionsFromRequest(req.GetOptions())

	msg, err = s.App.SendPay(ctx,
		msg.Payload, msg.AmtMsat, msg.DiscussionID, msg.PayReq, opts)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	resp, err := messageModelToSendMessageResponse(msg)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return resp, nil
}

// SubscribeMessages returns received messages on the provided grpc stream.
func (s *messageServiceServer) SubscribeMessages(_ *pb.SubscribeMessageRequest,
	srv pb.MessageService_SubscribeMessagesServer) error {

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	// Create a subscriber for received messages
	msgChannel, err := s.App.SubscribeMessages(ctx)
	if err != nil {
		return associateStatusCode(s.logError(
			fmt.Errorf("Client subscription failed: %w", err)))
	}

messageLoop:
	for {
		select {
		case <-ctx.Done():
			s.Log.Printf("Context cancelled")
			break messageLoop
		case msg, ok := <-msgChannel:
			if !ok {
				s.Log.Printf("Subscription channel closed.")
				break messageLoop
			}
			if msg.Error != nil {
				return associateStatusCode(s.logError(
					fmt.Errorf("message subscription error")))
			}

			// Forward received message to grpc stream
			resp, err := messageModelToSubscribeMessageResponse(msg.Message)
			if err != nil {
				return associateStatusCode(s.logError(err))
			}
			if err := srv.Send(resp); err != nil {
				return associateStatusCode(s.logError(err))
			}
		}
	}

	return nil
}

// NewMessageServiceServer initializes a new message service.
func NewMessageServiceServer(app *app.App) pb.MessageServiceServer {
	return &messageServiceServer{
		Log: slog.NewLogger("message-service"),
		App: app,
	}
}
