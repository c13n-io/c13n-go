package app

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/model"
	"github.com/c13n-io/c13n-go/store"
)

func (app *App) subscribePayments(ctx context.Context, lastPaymentIdx uint64) error {
	// Create subscription for payment updates
	paySubscription, err := app.LNManager.SubscribePaymentUpdates(ctx,
		lastPaymentIdx, defaultPaymentFilter)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-paySubscription:
			if !ok {
				return fmt.Errorf("payment subscription channel closed")
			}
			pmnt, err := update.Payment, update.Err
			if err != nil {
				return fmt.Errorf("payment update failed: %w", err)
			}

			payment := &model.Payment{
				PayerAddress: app.Self.Node.Address,
				Payment:      *pmnt,
			}
			switch dest, err := pmnt.GetDestination(); err {
			case nil:
				payment.PayeeAddress = dest.String()
			default:
				app.Log.WithError(err).Warn("could not retrieve payment destination")
			}

			// Store payment and notify topic
			if err := app.Database.AddPayments(payment); err != nil {
				var existsErr *store.AlreadyExistsError
				if !errors.As(err, &existsErr) {
					app.Log.WithError(err).Error("payment storage failed")
				}
			}

			if err := app.publishPayment(payment); err != nil {
				app.Log.WithError(err).Error("payment notification failed")
			}
		}
	}
}
