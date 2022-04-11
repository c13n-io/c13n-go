package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type discussionServiceServer struct {
	Log *slog.Logger

	App *app.App

	pb.UnimplementedDiscussionServiceServer
}

func (s *discussionServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// GetDiscussions returns information about all discussions
// over the provided grpc stream.
func (s *discussionServiceServer) GetDiscussions(_ *pb.GetDiscussionsRequest, srv pb.DiscussionService_GetDiscussionsServer) error {

	ctx := srv.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	discussions, err := s.App.GetDiscussions(ctx)
	if err != nil {
		return associateStatusCode(s.logError(err))
	}
	for _, disc := range discussions {
		discInfo, err := discussionModelToDiscussionInfo(&disc)
		if err != nil {
			return associateStatusCode(s.logError(err))
		}
		discResp := &pb.GetDiscussionsResponse{
			Discussion: discInfo,
		}
		if err := srv.Send(discResp); err != nil {
			return associateStatusCode(s.logError(err))
		}
	}
	return nil
}

// GetDiscussionHistoryByID returns previously exchanged messages
// associated with the discussion over the provided stream,
// respecting the pagination options parameter.
func (s *discussionServiceServer) GetDiscussionHistoryByID(req *pb.GetDiscussionHistoryByIDRequest, srv pb.DiscussionService_GetDiscussionHistoryByIDServer) error {

	ctx := srv.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	var pageOptions model.PageOptions
	if pageOpts := req.GetPageOptions(); pageOpts != nil {
		pageOptions = model.PageOptions{
			LastID:   pageOpts.GetLastId(),
			PageSize: uint64(pageOpts.GetPageSize()),
			Reverse:  pageOpts.GetReverse(),
		}
	}

	messages, err := s.App.GetDiscussionHistory(ctx, req.GetId(), pageOptions)
	if err != nil {
		return associateStatusCode(s.logError(err))
	}

	for _, message := range messages {
		msg, err := newMessage(&message)
		if err != nil {
			return associateStatusCode(s.logError(err))
		}

		resp := &pb.GetDiscussionHistoryResponse{
			Message: msg,
		}
		if err := srv.Send(resp); err != nil {
			return associateStatusCode(s.logError(err))
		}
	}
	return nil
}

// GetDiscussionStatistics returns statistics for the provided discussion.
func (s *discussionServiceServer) GetDiscussionStatistics(ctx context.Context, req *pb.GetDiscussionStatisticsRequest) (*pb.GetDiscussionStatisticsResponse, error) {
	stats, err := s.App.GetDiscussionStatistics(ctx, req.GetId())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.GetDiscussionStatisticsResponse{
		AmtMsatSent:      stats.AmtMsatSent,
		AmtMsatReceived:  stats.AmtMsatReceived,
		AmtMsatFees:      stats.AmtMsatFees,
		MessagesSent:     stats.MessagesSent,
		MessagesReceived: stats.MessagesReceived,
	}, nil
}

// AddDiscussion adds a discussion to the database.
// The discussion to be added is provided in the request.
func (s *discussionServiceServer) AddDiscussion(ctx context.Context, req *pb.AddDiscussionRequest) (*pb.AddDiscussionResponse, error) {
	discussion := discussionInfoToDiscussionModel(req.GetDiscussion())

	if len(discussion.Participants) < 1 {
		return nil, status.Error(codes.InvalidArgument,
			"Participant set empty for discussion")
	}
	// Disallow anonymous group discussions
	if len(discussion.Participants) > 1 && discussion.Options.Anonymous {
		return nil, status.Error(codes.InvalidArgument,
			"Anonymous group discussions are disallowed")
	}
	savedDiscussion, err := s.App.AddDiscussion(ctx, &discussion)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	responseDiscussion, err := discussionModelToDiscussionInfo(savedDiscussion)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.AddDiscussionResponse{
		Discussion: responseDiscussion,
	}, nil
}

// UpdateDiscussionLastRead updates the last read message id of a discussion.
func (s *discussionServiceServer) UpdateDiscussionLastRead(ctx context.Context,
	req *pb.UpdateDiscussionLastReadRequest) (*pb.UpdateDiscussionResponse, error) {

	discussionID, lastReadID := req.DiscussionId, req.LastReadMsgId

	if err := s.App.UpdateDiscussionLastRead(ctx, discussionID, lastReadID); err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.UpdateDiscussionResponse{}, nil
}

// RemoveDiscussion removes a discussion from the database,
// based on the id request field.
func (s *discussionServiceServer) RemoveDiscussion(ctx context.Context, req *pb.RemoveDiscussionRequest) (*pb.RemoveDiscussionResponse, error) {
	err := s.App.RemoveDiscussion(ctx, req.GetId())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.RemoveDiscussionResponse{}, nil
}

// Send sends a message over a payment.
// If a payment request is specified the discussion with the recipient is used,
// (creating it with default options if it does not exist).
func (s *discussionServiceServer) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	msg, payments, err := s.App.SendMessage(ctx,
		req.GetDiscussionId(), req.GetAmtMsat(), req.GetPayReq(),
		req.GetPayload(), messageOptionsFromRequest(req.GetOptions()))
	// SendMessage can partially succeed, in which case log the failures.
	if err != nil {
		rpcErr := associateStatusCode(s.logError(err))
		if msg == nil {
			return nil, rpcErr
		}
	}

	message, err := newSentMessage(msg, payments)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.SendResponse{
		SentMessage: message,
	}, nil
}

func newSentMessage(rawMsg *model.RawMessage, paymentList []*model.Payment) (*pb.Message, error) {
	payload, _, err := rawMsg.UnmarshalPayload()
	if err != nil {
		return nil, err
	}

	var sentAt, receivedAt *timestamppb.Timestamp
	var sentMsat uint64
	payments := make([]*pb.Payment, 0, len(paymentList))
	for _, p := range paymentList {
		payment, err := newPayment(p)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)

		if payment.GetState() == pb.PaymentState_PAYMENT_SUCCEEDED {
			sentMsat += payment.GetAmtMsat()
		}

		cTime := payment.GetCreatedTimestamp()
		rTime := payment.GetResolvedTimestamp()
		if sentAt.AsTime().After(cTime.AsTime()) || sentAt == nil {
			sentAt = cTime
		}
		if receivedAt.AsTime().Before(rTime.AsTime()) {
			receivedAt = rTime
		}
	}

	return &pb.Message{
		Id:                rawMsg.ID,
		DiscussionId:      rawMsg.DiscussionID,
		Sender:            rawMsg.Sender,
		SenderVerified:    rawMsg.SignatureVerified,
		Payload:           payload,
		AmtMsat:           int64(sentMsat),
		SentTimestamp:     sentAt,
		ReceivedTimestamp: receivedAt,
		LightningData: &pb.Message_Payments{
			Payments: &pb.Payments{
				Payments: payments,
			},
		},
	}, nil
}

func newMessage(aggregate *model.MessageAggregate) (*pb.Message, error) {
	if aggregate == nil {
		return nil, nil
	}

	raw := aggregate.RawMessage
	payload, _, err := raw.UnmarshalPayload()
	if err != nil {
		return nil, err
	}

	msg := &pb.Message{
		Id:             raw.ID,
		DiscussionId:   raw.DiscussionID,
		Sender:         raw.Sender,
		SenderVerified: raw.SignatureVerified,
		Payload:        payload,
	}

	var amtMsat uint64
	var sentAt, receivedAt *timestamppb.Timestamp
	switch {
	case raw.InvoiceSettleIndex != 0:
		invoice, err := invoiceModelToRPCInvoice(aggregate.Invoice)
		if err != nil {
			return nil, err
		}

		amtMsat = invoice.GetAmtPaidMsat()
		sentAt = invoice.GetCreatedTimestamp()
		receivedAt = invoice.GetSettledTimestamp()

		msg.LightningData = &pb.Message_Invoice{
			Invoice: invoice,
		}
	case len(raw.PaymentIndexes) != 0:
		payments := make([]*pb.Payment, len(aggregate.Payments))
		for i, pmnt := range aggregate.Payments {
			payment, err := newPayment(pmnt)
			if err != nil {
				return nil, err
			}

			if payment.GetState() == pb.PaymentState_PAYMENT_SUCCEEDED {
				amtMsat += payment.GetAmtMsat()
			}

			createdAt := payment.GetCreatedTimestamp()
			if sentAt.AsTime().After(createdAt.AsTime()) || sentAt == nil {
				sentAt = createdAt
			}
			resolvedAt := payment.GetResolvedTimestamp()
			if receivedAt.AsTime().Before(resolvedAt.AsTime()) {
				receivedAt = resolvedAt
			}

			payments[i] = payment
		}

		msg.LightningData = &pb.Message_Payments{
			Payments: &pb.Payments{
				Payments: payments,
			},
		}
	}

	msg.AmtMsat = int64(amtMsat)
	msg.SentTimestamp, msg.ReceivedTimestamp = sentAt, receivedAt

	return msg, nil
}

// NewDiscussionServiceServer initializes a new discussion service.
func NewDiscussionServiceServer(app *app.App) pb.DiscussionServiceServer {
	return &discussionServiceServer{
		Log: slog.NewLogger("message-service"),
		App: app,
	}
}
