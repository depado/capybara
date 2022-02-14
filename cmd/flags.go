package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

// addLoggerFlags adds support to configure the level of the logger.
func addLoggerFlags(c *cobra.Command) {
	c.PersistentFlags().String("log.level", "info", "one of debug, info, warn, error or fatal")
	c.PersistentFlags().String("log.type", "console", `one of "console" or "json"`)
	c.PersistentFlags().Bool("log.caller", true, "display the file and line where the call was made")
}

// addServerFlags adds support to configure the server.
func addServerFlags(c *cobra.Command) {
	c.PersistentFlags().String("server.host", "127.0.0.1", "host on which the server should listen")
	c.PersistentFlags().Int("server.port", 8080, "port on which the server should listen")
	c.PersistentFlags().String("server.tls.cert_path", "certs/server-cert.pem", "path to the server TLS certificate")
	c.PersistentFlags().String("server.tls.key_path", "certs/server-key.pem", "path to the certificate's private key")
	c.PersistentFlags().String("server.tls.type", "server", `one of "disable", "server", "mtls"`)
}

// addDatabaseFlags will add the database related flags and conf
func addDatabaseFlags(c *cobra.Command) {
	c.PersistentFlags().String("database.path", "capybara.db", "path to the database file to use")
	c.PersistentFlags().Duration("database.default_lock_ttl", 5*time.Minute, "default time to live for locks")
	c.PersistentFlags().Int("database.max_buckets_recursion", 3, "maximum recursion of buckets in database")
}

// addConfigurationFlag adds support to provide a configuration file on the
// command line.
func addConfigurationFlag(c *cobra.Command) {
	c.PersistentFlags().String("conf", "", "configuration file to use")
}
