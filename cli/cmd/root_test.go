package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/slog"
)

func TestMain(m *testing.M) {
	f, _ := ioutil.TempFile(os.TempDir(), "output-test_cmd-*.log")
	oldLogOut := slog.SetLogOutput(f)

	res := func() int {
		return m.Run()
	}()

	slog.SetLogOutput(oldLogOut)
	f.Close()

	os.Exit(res)
}

func TestInitConfig(t *testing.T) {
	// Create temp config file for test
	tmpConfigFile, err := ioutil.TempFile(os.TempDir(), "c13nTemp-*.yaml")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpConfigFile.Name())

	yamlLines := [][]byte{
		[]byte("log_level:     \"info\""),
		[]byte("server:"),
		[]byte("  address:        \"localhost:9998\""),
		[]byte("  graceful_shutdown_timeout: 11"),
		[]byte("  tls:"),
		[]byte("    cert_path: \"~/.c13n/service.pem\""),
		[]byte("    key_path: \"~/.c13n/service.key\""),
		[]byte("    extra_ips:"),
		[]byte("      - 10.10.10.10"),
		[]byte("      - 192.168.1.32"),
		[]byte("    extra_domains:"),
		[]byte("      - host_alias"),
		[]byte(""),
		[]byte("lnd:"),
		[]byte("  address:      \"localhost1:10009\""),
		[]byte("  tls_path:      \"~/.lnd/tls1.cert\""),
		[]byte("  macaroon_path: \"~/.lnd/data/chain/bitcoin/regtest/admin.macaroon1\""),
		[]byte("database:"),
		[]byte("  db_path:     \"./test.db\""),
		[]byte("  key_path:     \"./db_key\""),
	}

	yamlConfigText := bytes.Join(yamlLines, []byte("\n"))
	assert.NoError(t, err)

	writeLen, err := tmpConfigFile.Write(yamlConfigText)
	assert.Equal(t, len(yamlConfigText), writeLen)
	assert.NoError(t, err)
	// Close the file
	assert.NoError(t, tmpConfigFile.Close())

	//var lndConnectURL = ""
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	rootCmd.SetArgs([]string{"--config", tmpConfigFile.Name()})
	Execute()

	assert.Equal(t, "info", viper.GetString("log_level"))

	assert.Equal(t, "localhost:9998", viper.GetString("server.address"))
	assert.Equal(t, false, viper.GetBool("server.disable_tls"))
	assert.Equal(t, "~/.c13n/service.pem", viper.GetString("server.tls.cert_path"))
	assert.Equal(t, "~/.c13n/service.key", viper.GetString("server.tls.key_path"))
	assert.Equal(t, []string{"10.10.10.10", "192.168.1.32"},
		viper.GetStringSlice("server.tls.extra_ips"))
	assert.Equal(t, []string{"host_alias"},
		viper.GetStringSlice("server.tls.extra_domains"))
	assert.Equal(t, 11, viper.GetInt("server.graceful_shutdown_timeout"))

	assert.Equal(t, "localhost1:10009", viper.GetString("lnd.address"))
	assert.Equal(t, "~/.lnd/tls1.cert", viper.GetString("lnd.tls_path"))
	assert.Equal(t, "~/.lnd/data/chain/bitcoin/regtest/admin.macaroon1",
		viper.GetString("lnd.macaroon_path"))

	assert.Equal(t, "./test.db", viper.GetString("database.db_path"))
	assert.Equal(t, "./db_key", viper.GetString("database.key_path"))
}

/*
	This test validates the flag priority of cli. If a flag is defined
	through cli, the value in the config file is overwritten.
*/
func TestInitConfigPriority(t *testing.T) {
	// Create temp config file for test
	tmpConfigFile, err := ioutil.TempFile(os.TempDir(), "c13nTemp-*.yaml")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpConfigFile.Name())

	yamlLines := [][]byte{
		[]byte("log_level:     \"info\""),
		[]byte("server:"),
		[]byte("  address:        \"localhost:9998\""),
		[]byte("  graceful_shutdown_timeout: 11"),
		[]byte(""),
		[]byte("lnd:"),
		[]byte("  address:      \"localhost1:10009\""),
		[]byte("  tls_path:      \"~/.lnd/tls1.cert\""),
		[]byte("  macaroon_path: \"~/.lnd/data/chain/bitcoin/regtest/admin.macaroon1\""),
		[]byte("database:"),
		[]byte("  db_path:     \"./test.db\""),
		[]byte("  key_path:     \"./db_key\""),
	}

	yamlConfigText := bytes.Join(yamlLines, []byte("\n"))
	assert.NoError(t, err)

	writeLen, err := tmpConfigFile.Write(yamlConfigText)
	assert.Equal(t, len(yamlConfigText), writeLen)
	assert.NoError(t, err)
	// Close the file
	assert.NoError(t, tmpConfigFile.Close())

	//var lndConnectURL = ""
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	rootCmd.SetArgs([]string{"--config", tmpConfigFile.Name(),
		"--log-level", "debug",
		"--server-address", "random_host:5555",
		"--server-pwdhash", "bcrypt_hash",
		"--cert-path", "~/random/tls.cert",
		"--tls-extra-ip", "10.10.10.12",
		"--tls-extra-ip", "100.100.100.100",
		"--tls-extra-domain", "used_alias",
		"--key-path", "~/random/tls.key",
		"--graceful-shutdown-timeout", "12",
		"--lnd-address", "random_lnd_host:3333",
		"--lnd-tls-path", "tls-path",
		"--lnd-macaroon-path", "macaroon-path",
		"--db-path", "test-db-path",
		"--db-key-path", "db-encryption-key-path"})
	Execute()

	assert.Equal(t, "debug", viper.GetString("log_level"))

	assert.Equal(t, "random_host:5555", viper.GetString("server.address"))
	assert.Equal(t, "bcrypt_hash", viper.GetString("server.pwdhash"))
	assert.Equal(t, false, viper.GetBool("server.disable_tls"))
	assert.Equal(t, "~/random/tls.cert", viper.GetString("server.tls.cert_path"))
	assert.Equal(t, "~/random/tls.key", viper.GetString("server.tls.key_path"))
	assert.Equal(t, []string{"10.10.10.12", "100.100.100.100"},
		viper.GetStringSlice("server.tls.extra_ips"))
	assert.Equal(t, []string{"used_alias"},
		viper.GetStringSlice("server.tls.extra_domains"))
	assert.Equal(t, 12, viper.GetInt("server.graceful_shutdown_timeout"))

	assert.Equal(t, "random_lnd_host:3333", viper.GetString("lnd.address"))
	assert.Equal(t, "tls-path", viper.GetString("lnd.tls_path"))
	assert.Equal(t, "macaroon-path", viper.GetString("lnd.macaroon_path"))

	assert.Equal(t, "test-db-path", viper.GetString("database.db_path"))
	assert.Equal(t, "db-encryption-key-path", viper.GetString("database.key_path"))
}
