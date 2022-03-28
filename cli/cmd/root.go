package cmd

import (
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultCfgName = "c13n"
)

var cfgFile string
var lndConnectURL string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "c13n",
	Short: "Communication over Lightning",
	RunE:  Run,
}

func init() {
	initLogger()

	initCliFlags()
}

func initCliFlags() {
	cobra.OnInitialize(initConfig)

	// Disable sorting of flags in the usage string.
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.Flags().SortFlags = false

	rootFlags := rootCmd.PersistentFlags()

	rootFlags.StringVarP(&cfgFile, "config", "c", "",
		"Configuration file to be used for backend")

	rootFlags.StringP("log-level", "L", "debug",
		"Log level: (trace, debug, info, warn, error, fatal, panic)")
	_ = viper.BindPFlag("log_level", rootFlags.Lookup("log-level"))

	// APP flags
	rootFlags.Int64("default-fee-limit-msat", 3000,
		"Default fee limit for discussions in millisatoshi")
	_ = viper.BindPFlag("app.default_fee_limit_msat",
		rootFlags.Lookup("default-fee-limit-msat"))

	// RPC flags
	rootFlags.String("server-address", "localhost:9999",
		"Address to listen for incoming connections on")
	_ = viper.BindPFlag("server.address", rootFlags.Lookup("server-address"))
	rootFlags.String("cert-path", "",
		"Path of the TLS certificate used for server connections")
	_ = viper.BindPFlag("server.tls.cert_path", rootFlags.Lookup("cert-path"))
	rootFlags.String("key-path", "",
		"Path of the TLS private key used for server connections")
	_ = viper.BindPFlag("server.tls.key_path", rootFlags.Lookup("key-path"))
	rootFlags.StringSlice("tls-extra-ip", []string{},
		"Extra IPs to be added to the TLS certificate")
	_ = viper.BindPFlag("server.tls.extra_ips", rootFlags.Lookup("tls-extra-ip"))
	rootFlags.StringSlice("tls-extra-domain", []string{},
		"Extra domains to be added to the TLS certificate")
	_ = viper.BindPFlag("server.tls.extra_domains", rootFlags.Lookup("tls-extra-domain"))
	rootFlags.String("server-user", "", "Username for server connections")
	_ = viper.BindPFlag("server.user", rootFlags.Lookup("server-user"))
	rootFlags.String("server-passwd-hash", "", "Bcrypt password hash for server connections")
	_ = viper.BindPFlag("server.rpcpasswdhash", rootFlags.Lookup("server-passwd-hash"))
	rootFlags.Int("graceful-shutdown-timeout", 10,
		"Graceful shutdown timeout in seconds")
	_ = viper.BindPFlag("server.graceful_shutdown_timeout",
		rootFlags.Lookup("graceful-shutdown-timeout"))

	// LND flags
	rootFlags.String("lnd-address", "localhost:10009",
		"Address of the Lightning daemon")
	_ = viper.BindPFlag("lnd.address", rootFlags.Lookup("lnd-address"))
	rootFlags.String("lnd-tls-path", "",
		"Path of the Lightning daemon TLS certificate to use")
	_ = viper.BindPFlag("lnd.tls_path", rootFlags.Lookup("lnd-tls-path"))
	rootFlags.String("lnd-macaroon-path", "",
		"Path of the Lightning daemon macaroon file to use")
	_ = viper.BindPFlag("lnd.macaroon_path", rootFlags.Lookup("lnd-macaroon-path"))
	rootFlags.StringVarP(&lndConnectURL, "lndconnect", "l", "",
		"lndconnect URL to use for connection to the Lightning daemon")
	rootFlags.Int64("lnd-macaroon-timeout", 0,
		"Lifetime of transmitted daemon macaroon in seconds")
	_ = viper.BindPFlag("lnd.macaroon_timeout_secs",
		rootFlags.Lookup("lnd-macaroon-timeout"))
	rootFlags.String("lnd-macaroon-ip", "",
		"IP to lock the transmitted daemon macaroon to")
	_ = viper.BindPFlag("lnd.macaroon_ip", rootFlags.Lookup("lnd-macaroon-ip"))

	// DB flags
	rootFlags.String("db-path", "c13n.db",
		"Path of the database directory")
	_ = viper.BindPFlag("database.db_path", rootFlags.Lookup("db-path"))
	rootFlags.String("db-key-path", "",
		"Database encryption key of fixed length(16, 24 or 32 bytes)")
	_ = viper.BindPFlag("database.key_path", rootFlags.Lookup("db-key-path"))
}

// initConfig reads in config file and env variables if set.
func initConfig() {
	// Search config in home directory (if it can be located), or cwd (in that order).
	switch home, err := homedir.Dir(); err {
	case nil:
		viper.AddConfigPath(home)
	default:
		logger.WithError(err).Error("Could not locate HOME directory")
	}

	viper.AddConfigPath(".")
	// The following call set the default config file to "c13n.{type}".
	viper.SetConfigName(defaultCfgName)

	if cfgFile != "" {
		// Use config file from flag, overriding the defaults.
		viper.SetConfigFile(cfgFile)
	}

	// Read in matching environment variables.
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	switch err := viper.ReadInConfig(); err := err.(type) {
	case nil:
		logger.Infof("Using config file: %s", viper.ConfigFileUsed())
	case viper.ConfigFileNotFoundError:
		logger.WithError(err).Warn("Could not locate config file")
	default:
		logger.WithError(err).Warn("Configuration error")
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Warn("Execution terminated with error")
	}
}
