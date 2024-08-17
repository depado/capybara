package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/depado/capybara/cmd"
	"github.com/depado/capybara/database"
	"github.com/depado/capybara/server"
)

// Main function that will be executed from the root command.
func run() {
	conf, err := cmd.NewConf()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load configuration")
	}

	lg := cmd.NewLogger(conf)

	cdb, err := database.NewCapybaraDB(conf, lg)
	if err != nil {
		lg.Fatal().Err(err).Msg("unable to initialize database")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := cdb.Close()
		if err != nil {
			lg.Error().Err(err).Msg("closing database")
		}
		os.Exit(1)
	}()

	gs, err := server.NewGRPCServer(conf, lg, cdb)
	if err != nil {
		lg.Fatal().Err(err).Msg("unable to initialize grpc server")
	}
	server.Listen(conf, lg, gs)
}

// Main command that will be run when no other command is provided on the
// command-line.
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
