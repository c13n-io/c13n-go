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

	resp, err := newInvoice(inv)
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

	resp, err := newInvoice(inv)
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

// SubscribeInvoices returns a stream over which
// invoices that have reached a final state are received.
func (s *paymentServiceServer) SubscribeInvoices(_ *pb.SubscribeInvoicesRequest,
	srv pb.PaymentService_SubscribeInvoicesServer) error {

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	invChannel, err := s.App.SubscribeInvoices(ctx)
	if err != nil {
		return associateStatusCode(s.logError(
			fmt.Errorf("client subscription failed: %w", err)))
	}

invoiceLoop:
	for {
		select {
		case <-ctx.Done():
			s.Log.Printf("client subscription ended")
			break invoiceLoop
		case inv, ok := <-invChannel:
			if !ok {
				s.Log.Printf("subscription channel closed")
				break invoiceLoop
			}

			invoice, err := newInvoice(inv)
			if err != nil {
				return associateStatusCode(s.logError(err))
			}
			if err := srv.Send(invoice); err != nil {
				return associateStatusCode(s.logError(err))
			}
		}
	}

	return nil
}

// SubscribePayments returns a stream over which
// notifications of finished payments are received.
func (s *paymentServiceServer) SubscribePayments(_ *pb.SubscribePaymentsRequest,
	srv pb.PaymentService_SubscribePaymentsServer) error {

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	payChannel, err := s.App.SubscribePayments(ctx)
	if err != nil {
		return associateStatusCode(s.logError(
			fmt.Errorf("client subscription failed: %w", err)))
	}

paymentLoop:
	for {
		select {
		case <-ctx.Done():
			s.Log.Printf("client subscription ended")
			break paymentLoop
		case pmnt, ok := <-payChannel:
			if !ok {
				s.Log.Printf("subscription channel closed")
				break paymentLoop
			}

			payment, err := newPayment(pmnt)
			if err != nil {
				return associateStatusCode(s.logError(err))
			}
			if err := srv.Send(payment); err != nil {
				return associateStatusCode(s.logError(err))
			}
		}
	}

	return nil
}

func newPayment(payment *model.Payment) (*pb.Payment, error) {
	var err error
	var createdTime, resolvedTime *timestamppb.Timestamp
	htlcs := make([]*pb.PaymentHTLC, len(payment.Htlcs))
	for i, htlc := range payment.Htlcs {
		htlcs[i], err = newPaymentHTLC(htlc)
		if err != nil {
			return nil, err
		}

		if htlcs[i].GetResolveTimestamp().AsTime().After(resolvedTime.AsTime()) || resolvedTime == nil {
			resolvedTime = htlcs[i].GetResolveTimestamp()
		}
	}

	if createdTime, err = newProtoTimestamp(time.Unix(0, payment.CreationTimeNs)); err != nil {
		return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
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

	var hPreimage string
	if h.Preimage != nil {
		preimage, err := lntypes.MakePreimage(h.Preimage)
		if err != nil {
			return nil, fmt.Errorf("marshal error: invalid preimage: %v", err)
		}
		hPreimage = preimage.String()
	}

	return &pb.PaymentHTLC{
		Route:            route,
		AttemptTimestamp: attemptTime,
		ResolveTimestamp: resolveTime,
		State:            state,
		Preimage:         hPreimage,
	}, nil

}

var (
	invoiceStateMap = map[lnchat.InvoiceState]pb.InvoiceState{
		lnchat.InvoiceOPEN:      pb.InvoiceState_INVOICE_OPEN,
		lnchat.InvoiceACCEPTED:  pb.InvoiceState_INVOICE_ACCEPTED,
		lnchat.InvoiceSETTLED:   pb.InvoiceState_INVOICE_SETTLED,
		lnchat.InvoiceCANCELLED: pb.InvoiceState_INVOICE_CANCELLED,
	}
	invoiceHTLCStateMap = map[lnrpc.InvoiceHTLCState]pb.InvoiceHTLCState{
		lnrpc.InvoiceHTLCState_ACCEPTED: pb.InvoiceHTLCState_INVOICE_HTLC_ACCEPTED,
		lnrpc.InvoiceHTLCState_SETTLED:  pb.InvoiceHTLCState_INVOICE_HTLC_SETTLED,
		lnrpc.InvoiceHTLCState_CANCELED: pb.InvoiceHTLCState_INVOICE_HTLC_CANCELLED,
	}
)

func newInvoice(invoice *model.Invoice) (*pb.Invoice, error) {
	var err error
	var created, settled *timestamppb.Timestamp
	if invoice.CreatedTimeSec > 0 {
		ts := time.Unix(invoice.CreatedTimeSec, 0)
		if created, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}
	if invoice.SettleTimeSec > 0 {
		ts := time.Unix(invoice.SettleTimeSec, 0)
		if settled, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}

	preimage, err := lntypes.MakePreimage(invoice.Preimage)
	if err != nil {
		return nil, fmt.Errorf("marshal error: invalid preimage: %v", err)
	}

	hints, err := newInvoiceHints(invoice.RouteHints)
	if err != nil {
		return nil, fmt.Errorf("marshal error: route hints error: %v", err)
	}

	htlcs, err := newInvoiceHTLCs(invoice.Htlcs)
	if err != nil {
		return nil, fmt.Errorf("marshal error: htlc error: %v", err)
	}

	return &pb.Invoice{
		Memo:             invoice.Memo,
		Hash:             invoice.Hash,
		Preimage:         preimage.String(),
		PaymentRequest:   invoice.PaymentRequest,
		ValueMsat:        uint64(invoice.Value.Msat()),
		AmtPaidMsat:      uint64(invoice.AmtPaid.Msat()),
		CreatedTimestamp: created,
		SettledTimestamp: settled,
		Expiry:           invoice.Expiry,
		Private:          invoice.Private,
		RouteHints:       hints,
		State:            invoiceStateMap[invoice.State],
		AddIndex:         invoice.AddIndex,
		SettleIndex:      invoice.SettleIndex,
		InvoiceHtlcs:     htlcs,
	}, nil
}

func newInvoiceHTLCs(htlcs []lnchat.InvoiceHTLC) ([]*pb.InvoiceHTLC, error) {
	res := make([]*pb.InvoiceHTLC, len(htlcs))
	for i, htlc := range htlcs {
		var err error
		var accept, resolve *timestamppb.Timestamp
		if htlc.AcceptTimeSec > 0 {
			ts := time.Unix(htlc.AcceptTimeSec, 0)
			if accept, err = newProtoTimestamp(ts); err != nil {
				return nil, fmt.Errorf("marshal error: "+
					"invalid timestamp: %v", err)
			}
		}
		if htlc.ResolveTimeSec > 0 {
			ts := time.Unix(htlc.ResolveTimeSec, 0)
			if resolve, err = newProtoTimestamp(ts); err != nil {
				return nil, fmt.Errorf("marshal error: "+
					"invalid timestamp: %v", err)
			}
		}

		res[i] = &pb.InvoiceHTLC{
			ChanId:           htlc.ChanID,
			AmtMsat:          uint64(htlc.Amount.Msat()),
			State:            invoiceHTLCStateMap[htlc.State],
			AcceptTimestamp:  accept,
			ResolveTimestamp: resolve,
			ExpiryHeight:     htlc.ExpiryHeight,
		}
	}

	return res, nil
}

func newInvoiceHints(hints []lnchat.RouteHint) ([]*pb.RouteHint, error) {
	res := make([]*pb.RouteHint, len(hints))
	for i, hint := range hints {
		hintHops := make([]*pb.HopHint, len(hint.HopHints))

		for j, hop := range hint.HopHints {
			hintHops[j] = &pb.HopHint{
				Pubkey:          hop.NodeID.String(),
				ChanId:          hop.ChanID,
				FeeBaseMsat:     hop.FeeBaseMsat,
				FeeRate:         hop.FeeRate,
				CltvExpiryDelta: hop.CltvExpiryDelta,
			}
		}

		res[i] = &pb.RouteHint{
			HopHints: hintHops,
		}
	}

	return res, nil
}

// NewPaymentServiceServer initializes a new payment service.
func NewPaymentServiceServer(app *app.App) pb.PaymentServiceServer {
	return &paymentServiceServer{
		Log: slog.NewLogger("payment-service"),
		App: app,
	}
}
