package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type paymentServiceServer struct {
	Log *slog.Logger

	App *app.App

	pb.UnimplementedPaymentServiceServer
}

func (s *paymentServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// CreateInvoice creates and returns an invoice for the specified amount
// with the specified memo and expiry time.
func (s *paymentServiceServer) CreateInvoice(ctx context.Context, req *pb.CreateInvoiceRequest) (*pb.CreateInvoiceResponse, error) {

	inv, err := s.App.CreateInvoice(ctx,
		req.Memo, int64(req.AmtMsat), req.Expiry, req.Private)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	resp, err := invoiceModelToRPCInvoice(inv)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.CreateInvoiceResponse{
		Invoice: resp,
	}, nil
}

// LookupInvoice retrieves an invoice and returns it.
func (s *paymentServiceServer) LookupInvoice(ctx context.Context, req *pb.LookupInvoiceRequest) (*pb.LookupInvoiceResponse, error) {
	inv, err := s.App.LookupInvoice(ctx, req.GetPayReq())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	resp, err := invoiceModelToRPCInvoice(inv)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.LookupInvoiceResponse{
		Invoice: resp,
	}, nil
}

// Pay performs a payment and returns it.
func (s *paymentServiceServer) Pay(ctx context.Context,
	req *pb.PayRequest) (*pb.PayResponse, error) {

	paymentOptions := app.DefaultPaymentOptions
	paymentOptions.FeeLimitMsat = req.Options.GetFeeLimitMsat()

	payment, err := s.App.SendPayment(ctx,
		req.GetAddress(), int64(req.GetAmtMsat()), req.GetPayReq(),
		paymentOptions, nil)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	resp, err := newPayment(payment)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.PayResponse{
		Payment: resp,
	}, nil
}

func newPayment(payment *model.Payment) (*pb.Payment, error) {
	var err error
	var createdTime, resolvedTime *timestamppb.Timestamp
	resolvedTimeNs := int64(0)
	// Assign resolvedTimeNs to the latest succeeded htlc resolve timestamp
	for _, h := range payment.Htlcs {
		if h.Status == lnrpc.HTLCAttempt_SUCCEEDED && h.ResolveTimeNs > resolvedTimeNs {
			resolvedTimeNs = h.ResolveTimeNs
		}
	}

	htlcs := make([]*pb.PaymentHTLC, len(payment.Htlcs))
	for i := range htlcs {
		htlcs[i], err = newPaymentHTLC(payment.Htlcs[i])
		if err != nil {
			return nil, err
		}

		if payment.Htlcs[i].ResolveTimeNs > resolvedTimeNs {
			resolvedTimeNs = payment.Htlcs[i].ResolveTimeNs
		}
	}

	if resolvedTimeNs > 0 {
		ts := time.Unix(0, resolvedTimeNs)
		if resolvedTime, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}
	if payment.CreationTimeNs > 0 {
		ts := time.Unix(0, payment.CreationTimeNs)
		if createdTime, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}

	var state pb.PaymentState
	switch payment.Status {
	case lnchat.PaymentUNKNOWN:
		state = pb.PaymentState_PAYMENT_UNKNOWN
	case lnchat.PaymentINFLIGHT:
		state = pb.PaymentState_PAYMENT_INFLIGHT
	case lnchat.PaymentSUCCEEDED:
		state = pb.PaymentState_PAYMENT_SUCCEEDED
	case lnchat.PaymentFAILED:
		state = pb.PaymentState_PAYMENT_FAILED
	default:
		return nil, fmt.Errorf("marshal error: invalid payment state: %v", err)
	}

	return &pb.Payment{
		Hash:              payment.Hash,
		Preimage:          payment.Preimage,
		AmtMsat:           uint64(payment.Value.Msat()),
		CreatedTimestamp:  createdTime,
		ResolvedTimestamp: resolvedTime,
		PayReq:            payment.PaymentRequest,
		State:             state,
		PaymentIndex:      payment.PaymentIndex,
		HTLCs:             htlcs,
	}, nil
}

func newPaymentHTLC(h lnchat.HTLCAttempt) (*pb.PaymentHTLC, error) {
	hops := make([]*pb.PaymentHop, len(h.Route.Hops))
	for i, hop := range h.Route.Hops {
		hops[i] = &pb.PaymentHop{
			ChanId:           hop.ChannelID,
			HopAddress:       hop.NodeID.String(),
			AmtToForwardMsat: hop.AmtToForward.Msat(),
			FeeMsat:          int64(hop.Fees.Msat()),
		}
	}

	route := &pb.PaymentRoute{
		Hops:          hops,
		TotalTimelock: h.Route.TimeLock,
		RouteAmtMsat:  h.Route.Amt.Msat(),
		RouteFeesMsat: h.Route.Fees.Msat(),
	}

	var attemptTime, resolveTime *timestamppb.Timestamp
	var err error
	if h.AttemptTimeNs > 0 {
		ts := time.Unix(0, h.AttemptTimeNs)
		if attemptTime, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}
	if h.ResolveTimeNs > 0 {
		ts := time.Unix(0, h.ResolveTimeNs)
		if resolveTime, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}

	var state pb.HTLCState
	switch h.Status {
	case lnrpc.HTLCAttempt_IN_FLIGHT:
		state = pb.HTLCState_HTLC_IN_FLIGHT
	case lnrpc.HTLCAttempt_SUCCEEDED:
		state = pb.HTLCState_HTLC_SUCCEEDED
	case lnrpc.HTLCAttempt_FAILED:
		state = pb.HTLCState_HTLC_FAILED
	}

	preimage, err := lntypes.MakePreimage(h.Preimage)
	if err != nil {
		return nil, fmt.Errorf("marshal error: invalid preimage: %v", err)
	}

	return &pb.PaymentHTLC{
		Route:            route,
		AttemptTimestamp: attemptTime,
		ResolveTimestamp: resolveTime,
		State:            state,
		Preimage:         preimage.String(),
	}, nil

}

// NewPaymentServiceServer initializes a new payment service.
func NewPaymentServiceServer(app *app.App) pb.PaymentServiceServer {
	return &paymentServiceServer{
		Log: slog.NewLogger("payment-service"),
		App: app,
	}
}
