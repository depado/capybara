package cmd

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AddLoggerFlags adds support to configure the level of the logger.
func AddLoggerFlags(c *cobra.Command) {
	c.PersistentFlags().String("log.level", "info", "one of debug, info, warn, error or fatal")
	c.PersistentFlags().String("log.type", "console", `one of "console" or "json"`)
	c.PersistentFlags().Bool("log.caller", true, "display the file and line where the call was made")
}

// AddServerFlags adds support to configure the server.
func AddServerFlags(c *cobra.Command) {
	c.PersistentFlags().String("server.host", "127.0.0.1", "host on which the server should listen")
	c.PersistentFlags().Int("server.port", 8080, "port on which the server should listen")
}

// AddDatabaseFlags will add the database related flags and conf
func AddDatabaseFlags(c *cobra.Command) {
	c.PersistentFlags().String("database.path", "capybara.db", "path to the database file to use")
	c.PersistentFlags().Duration("database.default_lock_ttl", 5*time.Minute, "default time to live for locks")
	c.PersistentFlags().Int("database.max_buckets_recursion", 3, "maximum recursion of buckets in database")
}

// AddConfigurationFlag adds support to provide a configuration file on the
// command line.
func AddConfigurationFlag(c *cobra.Command) {
	c.PersistentFlags().String("conf", "", "configuration file to use")
}

// AddAllFlags will add all the flags provided in this package to the provided
// command and will bind those flags with viper.
func AddAllFlags(c *cobra.Command) {
	AddConfigurationFlag(c)
	AddLoggerFlags(c)
	AddServerFlags(c)
	AddDatabaseFlags(c)

	if err := viper.BindPFlags(c.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("unable to bind flags")
	}
}
