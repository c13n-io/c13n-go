package app

import (
	"context"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// CreateInvoice creates an invoice and returns it.
func (app *App) CreateInvoice(ctx context.Context, memo string,
	amtMsat int64, expiry int64, private bool) (*model.Invoice, error) {

	inv, err := app.LNManager.CreateInvoice(ctx, memo,
		lnchat.NewAmount(amtMsat), expiry, private)
	if err != nil {
		return nil, err
	}

	return &model.Invoice{
		CreatorAddress: app.Self.Node.Address,
		Invoice:        *inv,
	}, nil
}

// LookupInvoice retrieves an invoice and returns it.
func (app *App) LookupInvoice(ctx context.Context, payReq string) (*model.Invoice, error) {
	res, err := app.LNManager.DecodePayReq(ctx, payReq)
	if err != nil {
		return nil, err
	}

	inv, err := app.LNManager.LookupInvoice(ctx, res.Hash)
	if err != nil {
		return nil, err
	}

	return &model.Invoice{
		CreatorAddress: app.Self.Node.Address,
		Invoice:        *inv,
	}, nil
}

// GetInvoices retrieves stored invoices.
func (app *App) GetInvoices(_ context.Context, pageOpts model.PageOptions) ([]*model.Invoice, error) {
	invoices, err := app.Database.GetInvoices(pageOpts)
	if err != nil {
		return nil, newErrorf(err, "could not retrieve invoices")
	}

	return invoices, nil
}
