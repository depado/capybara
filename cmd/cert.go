package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Used to interact with capybara cert generator for grpc",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var certGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a new certificate ready to use",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := NewConf()
		if err != nil {
			l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
			l.Fatal().Err(err).Msg("unable to parse configuration")
		}
		l := NewLogger(conf)
		if err := GenerateServerCertEd(conf, l, true); err != nil {
			l.Fatal().Err(err).Msg("unable to generate server cert")
		}
	},
}

var certCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check certificates",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
