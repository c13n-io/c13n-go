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
