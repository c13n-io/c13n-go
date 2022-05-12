package app

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/pkg/errors"
	"gopkg.in/tomb.v2"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/slog"
	"github.com/c13n-io/c13n-go/store"
)

// App defines the c13n application logic.
type App struct {
	Log *slog.Logger

	Self lnchat.SelfInfo

	LNManager lnchat.LightManager
	Database  store.Database

	bus *gochannel.GoChannel

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

func backoff(n int) time.Duration {
	startBackoff, maxCeilOffset := 5., 595.

	// backoff duration follows a modified sigmoid function
	exp := math.Exp2(float64(n - 10))
	nextBackoff := math.Trunc(startBackoff + maxCeilOffset*(exp)/(exp+1))

	return time.Duration(int64(nextBackoff)) * time.Second
}

func runGo(t *tomb.Tomb, log *slog.Logger, name string, run func(context.Context) error) {
	t.Go(func() error {
		for failCount := 0; t.Alive(); {
			// tomb.Tomb returns the previously provided parent (or a background context)
			// if called with nil context
			if err := run(t.Context(nil)); err != nil { //nolint:staticcheck // See above
				log.WithError(err).Warnf("%s terminated erroneously", name)
				switch {
				case errors.Is(err, lnchat.ErrNetworkUnavailable):
					failCount++
				default:
					failCount = 0
				}
			}

			boff := backoff(failCount)
			log.Infof("retrying %s after %s (attempt %d)", name, boff, failCount+1)
			select {
			case <-time.After(boff):
			case <-t.Dying():
			}
		}

		log.Infof("%s terminated", name)
		return nil
	})
}

// Init performs any initializations needed at the logic layer, and also
// opens all persistent subscriptions to the Lightning daemon.
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

	// Initialize bus for publishing events
	app.Log.Info("Creating pubsub bus")
	app.bus = gochannel.NewGoChannel(
		gochannel.Config{},
		slog.NewWLogger("watermill"),
	)

	// Run the subscriptions as separate goroutines, listening for events
	// and publishing them on the proper bus topic.
	app.Tomb, _ = tomb.WithContext(ctx)
	runGo(app.Tomb, app.Log, "invoice subscription", func(ctx context.Context) error {
		lastInvoiceSettleIdx, err := app.Database.GetLastInvoiceIndex()
		if err != nil {
			return fmt.Errorf("could not retrieve last known invoice: %v", err)
		}

		return app.subscribeInvoices(ctx, lastInvoiceSettleIdx)
	})

	return nil
}

// Cleanup performs cleanup and shutdown.
func (app *App) Cleanup() (err error) {
	if app.Tomb != nil {
		app.Tomb.Kill(nil)
	}

	app.Log.Info("Closing pubsub bus")
	if app.bus != nil {
		if err = app.bus.Close(); err != nil {
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
