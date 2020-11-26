package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/Depado/capybara/cmd"
	"github.com/Depado/capybara/database"
	"github.com/Depado/capybara/server"
)

// Build number and versions injected at compile time, set yours
var (
	Version = "unknown"
	Build   = "unknown"
)

// Main command that will be run when no other command is provided on the
// command-line
var rootCmd = &cobra.Command{
	Use: "capybara",
	Run: func(cmd *cobra.Command, args []string) { run() },
}

// Version command that will display the build number and version (if any)
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show build and version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Build: %s\nVersion: %s\n", Build, Version)
	},
}

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

func main() {
	// Initialize Cobra and Viper
	// cobra.OnInitialize(cmd.Initialize)
	cmd.AddAllFlags(rootCmd)
	rootCmd.AddCommand(versionCmd)

	// Run the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("unable to start")
	}
}
