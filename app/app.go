package app

import (
	"context"
	"math"
	"time"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/pkg/errors"
	"gopkg.in/tomb.v2"

	"github.com/c13n-io/c13n-backend/lnchat"
	"github.com/c13n-io/c13n-backend/slog"
	"github.com/c13n-io/c13n-backend/store"
)

// App defines the c13n application logic.
type App struct {
	Log *slog.Logger

	Self lnchat.SelfInfo

	LNManager lnchat.LightManager
	Database  store.Database

	PubSubBus *gochannel.GoChannel

	Tomb *tomb.Tomb
}

// New creates a new app instance.
func New(lnChat lnchat.LightManager, database store.Database,
	options ...func(*App) error) (*App, error) {

	app := &App{
		Log:       slog.NewLogger("app"),
		LNManager: lnChat,
		Database:  database,
	}

	for _, option := range options {
		if err := option(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

// WithDefaultFeeLimitMsat sets the FeeLimitMsat default value
// for the app instance.
// The FeeLimitMsat default value is the default maximum fee
// used for sending a message.
func WithDefaultFeeLimitMsat(defaultLimitMsat int64) func(*App) error {
	return func(app *App) error {
		DefaultOptions.FeeLimitMsat = defaultLimitMsat
		return nil
	}
}

func subscriptionBackoffFn(n int) time.Duration {
	startBackoff, maxCeilOffset := 5., 595.

	// Backoff follows a modified sigmoid function
	exp := math.Exp2(float64(n - 10))
	nextBackoff := math.Trunc(startBackoff + maxCeilOffset*(exp)/(exp+1))

	return time.Duration(int64(nextBackoff)) * time.Second
}

// Init performs any initializations needed at the logic layer, and also
// opens a persistent publisher listening for received messages from
// the Lightning daemon.
func (app *App) Init(ctx context.Context, infoTimeoutSecs uint) error {
	var err error

	// Retrieve underlying node identity
	ctxt, cancel := context.WithTimeout(ctx,
		time.Duration(infoTimeoutSecs)*time.Second)
	defer cancel()
	app.Self, err = app.LNManager.GetSelfInfo(ctxt)
	if err != nil {
		return newErrorf(err, "GetSelfInfo")
	}

	// Initialize GoChannel for publishing received messages
	app.Log.Info("Creating pubsub bus")
	app.PubSubBus = gochannel.NewGoChannel(
		gochannel.Config{},
		slog.NewWLogger("watermill"),
	)

	// Run the message subscription publisher as a separate goroutine,
	// listening for incoming messages and publishing them
	// under the ReceiveTopic on App.PubSubBus
	var subscriptionCtx context.Context
	app.Tomb, subscriptionCtx = tomb.WithContext(ctx)
	app.Tomb.Go(func() error {
		// Until the app is requested to terminate,
		// recreate invoice subscription each time it terminates.
		for failedCount := 0; app.Tomb.Alive(); {
			// Retrieve last known invoice
			lastInvoiceSettleIdx, err := app.Database.GetLastInvoiceIndex()
			if err != nil {
				app.Log.WithError(err).Warn("could not retrieve last known invoice")
				continue
			}

			err = app.subscribeInvoices(subscriptionCtx, lastInvoiceSettleIdx)
			switch {
			case err != nil:
				app.Log.WithError(err).Warn("invoice subscription terminated erroneously")
				// In case of disconnection, increment backoff.
				if errors.Is(err, lnchat.ErrNetworkUnavailable) {
					failedCount++
					break
				}
				fallthrough
			default:
				failedCount = 0
			}

			backoffInterval := subscriptionBackoffFn(failedCount)
			app.Log.Infof("retrying subscription after %s "+
				"(attempt %d)", backoffInterval, failedCount+1)

			// Wait for backoff to elapse, while also
			// listening for normal termination.
			select {
			case <-time.After(backoffInterval):
			case <-app.Tomb.Dying():
			}
		}
		app.Log.Info("invoice subscription terminated")
		return nil
	})

	return nil
}

// Cleanup performs cleanup and shutdown.
func (app *App) Cleanup() (err error) {
	if app.Tomb != nil {
		app.Tomb.Kill(nil)
	}

	app.Log.Info("Closing pubsub bus")
	if app.PubSubBus != nil {
		if err = app.PubSubBus.Close(); err != nil {
			app.Log.WithError(err).Warn("PubSub bus close failed")
		}
	}

	app.Log.Info("Closing Lightning manager")
	if err = app.LNManager.Close(); err != nil {
		app.Log.WithError(err).Warn("Lightning manager shutdown failed")
	}

	if app.Tomb != nil {
		if err = app.Tomb.Wait(); err != nil {
			app.Log.WithError(err).Warn("Message subscription termination failed")
		}
	}

	app.Log.Info("Closing database")
	if err := app.Database.Close(); err != nil {
		app.Log.WithError(err).Warn("Database close failed")
	}

	return
}
