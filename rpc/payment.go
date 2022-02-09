package rpc

import (
	"context"

	"github.com/c13n-io/c13n-go/app"
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

// NewPaymentServiceServer initializes a new payment service.
func NewPaymentServiceServer(app *app.App) pb.PaymentServiceServer {
	return &paymentServiceServer{
		Log: slog.NewLogger("payment-service"),
		App: app,
	}
}
