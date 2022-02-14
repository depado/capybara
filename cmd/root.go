package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Setup(com *cobra.Command) {
	// Add and parse flags
	addConfigurationFlag(com)
	addLoggerFlags(com)
	addServerFlags(com)
	addDatabaseFlags(com)

	// Bind flags
	if err := viper.BindPFlags(com.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("unable to bind flags")
	}

	// Add version command
	com.AddCommand(versionCmd)

	// Setup cert command
	certCmd.AddCommand(certCheckCmd)
	certCmd.AddCommand(certGenCmd)

	// Add cert command
	com.AddCommand(certCmd)
}
