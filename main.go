package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/Depado/capybara/cmd"
	"github.com/Depado/capybara/database"
	"github.com/Depado/capybara/server"
)

// Main function that will be executed from the root command
func run() {
	fx.New(
		fx.NopLogger,
		fx.Provide(
			cmd.NewConf, cmd.NewLogger,
			database.NewCapybaraDB,
			server.NewGRPCServer,
		),
		fx.Invoke(server.Listen),
	).Run()
}

// Main command that will be run when no other command is provided on the
// command-line
var rootCmd = &cobra.Command{
	Use: "capybara",
	Run: func(cmd *cobra.Command, args []string) { run() },
}

func main() {
	// Setup command line
	cmd.Setup(rootCmd)

	// Run the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("unable to start")
	}
}
