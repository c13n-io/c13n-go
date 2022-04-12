package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
	"github.com/c13n-io/c13n-go/rpc"
	"github.com/c13n-io/c13n-go/slog"
	"github.com/c13n-io/c13n-go/store"
)

var (
	logger *slog.Logger
	server *rpc.Server
)

func initLogger() {
	// Create temporary cmd logger
	logger = slog.NewLogger("cmd")
}

// Run initializes the configuration and starts the application.
func Run(_ *cobra.Command, _ []string) error {
	// Set the default log level
	logLevel := viper.GetString("log_level")
	if err := slog.SetLogLevel(logLevel); err != nil {
		logger.WithError(err).
			Errorf("Could not set log level to %q", logLevel)
		return err
	}
	// Recreate cmd logger after the log level initialization
	logger = slog.NewLogger("cmd")

	// Open database encryption file
	dbMasterKey, err := ioutil.ReadFile(viper.GetString("database.key_path"))
	if err != nil {
		logger.WithError(err).Error("Could not read database encryption key file")
		return err
	}

	dbKeyLen := len(dbMasterKey)
	if dbKeyLen != 16 && dbKeyLen != 32 && dbKeyLen != 64 {
		logger.WithError(err).Error("Database encryption key not of standard size (16,32,64 bytes)")
		return err
	}

	// Initialize database
	db, err := store.New(viper.GetString("database.db_path"), store.WithBadgerOption(
		func(o badger.Options) badger.Options {
			return o.WithEncryptionKey(dbMasterKey).WithIndexCacheSize(1 << 20)
		}),
	)
	if err != nil {
		logger.WithError(err).Error("Could not create database")
		return err
	}

	// Initialize chat service
	var creds lnconnect.Credentials

	macConstraints := lnchat.MacaroonConstraints{
		Timeout: viper.GetInt64("lnd.macaroon_timeout_secs"),
		IPLock:  viper.GetString("lnd.macaroon_ip"),
	}

	if viper.GetString("lndconnect") != "" {
		creds, err = lnchat.NewCredentialsFromURL(
			viper.GetString("lndconnect"),
			macConstraints,
		)
	} else {
		creds, err = lnchat.NewCredentials(
			viper.GetString("lnd.address"),
			viper.GetString("lnd.tls_path"),
			viper.GetString("lnd.macaroon_path"),
			macConstraints,
		)
	}
	if err != nil {
		logger.WithError(err).Error("Could not create credentials")
		return err
	}

	lnchatMgr, err := lnchat.New(creds)
	if err != nil {
		logger.WithError(err).Error("Could not initialize lnchat service")
		return err
	}

	ctxb := context.Background()
	globalCtx, globalCancel := context.WithCancel(ctxb)
	defer globalCancel()

	// Initialize application
	var appOpts []func(*app.App) error

	defaultFeeLimitMsat := viper.GetInt64("app.default_fee_limit_msat")
	if defaultFeeLimitMsat != 0 {
		appOpts = append(appOpts, app.WithDefaultFeeLimitMsat(defaultFeeLimitMsat))
	}
	application, err := app.New(lnchatMgr, db, appOpts...)
	if err != nil {
		logger.WithError(err).Error("Could not create application")
		return err
	}

	if err := application.Init(globalCtx, 15); err != nil {
		logger.WithError(err).Error("Could not initialize application")
		return err
	}

	// Initialize server
	var srvOpts []func(*rpc.Server) error
	if viper.IsSet("server.tls.cert_path") && viper.IsSet("server.tls.key_path") {
		srvOpts = append(srvOpts, rpc.WithTLS(
			viper.GetString("server.tls.cert_path"),
			viper.GetString("server.tls.key_path"),
		))
	}
	if viper.IsSet("server.user") && viper.IsSet("server.pwdhash") {
		srvOpts = append(srvOpts, rpc.WithBasicAuth(
			viper.GetString("server.user"),
			viper.GetString("server.pwdhash"),
		))
	}
	srvAddress := viper.GetString("server.address")
	server, err = rpc.New(srvAddress, application, srvOpts...)
	if err != nil {
		logger.WithError(err).Error("Could not initialize server")
		return err
	}

	// Shutdown on interrupt
	terminationCh := make(chan interface{})
	go waitForTermination(terminationCh,
		time.Duration(viper.GetInt("server.graceful_shutdown_timeout"))*time.Second)

	logger.Infof("Starting server on %s", srvAddress)

	// Run server
	if err := server.Serve(server.Listener); err != nil {
		logger.WithError(err).Error("Fatal server error during Serve")
		return err
	}

	<-terminationCh

	logger.Info("THE END")
	return nil
}

func waitForTermination(terminationCh chan<- interface{}, gracePeriodTimeout time.Duration) {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)

	<-interruptCh
	logger.Info("Received shutdown signal.")

	// Try to terminate gracefully
	graceWaitCh := make(chan struct{})
	go func() {
		//graceWaitCh <- true
		server.GracefulStop()
		close(graceWaitCh)
	}()

	// Stop the server
	logger.Infof("Waiting %v for graceful termination.", gracePeriodTimeout)
	select {
	case <-graceWaitCh:
	case <-time.After(gracePeriodTimeout):
		server.Stop()
	}

	if err := server.Cleanup(); err != nil {
		logger.WithError(err).Error("Error generated during cleanup")
	}

	close(terminationCh)
}
