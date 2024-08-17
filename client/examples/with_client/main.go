package main

import (
	"os"

	capybara "github.com/depado/capybara/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	lock := "hi"

	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize the client
	capy, err := capybara.NewClient("127.0.0.1:8080", capybara.ClientOpts{
		Token:    "valid-token",
		CertPath: "certs/ca-cert.pem",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initialize capybara client")
	}
	// Close the connection
	defer func() {
		if err := capy.Close(); err != nil {
			log.Error().Err(err).Msg("unable to close connection")
		}

		log.Info().Str("who", capy.WhoAmI()).Msg("connection closed")
	}()

	// Claim given lock
	if _, err = capy.ClaimLock(lock); err != nil {
		log.Error().Err(err).Msg("unable to claim lock")
		return
	}

	log.Info().Str("lock", lock).Msg("claimed lock")
	// Defer the lock release
	defer func() {
		if err := capy.ReleaseLock(lock); err != nil {
			log.Error().Err(err).Msg("unable to release lock")
			return
		}

		log.Info().Str("lock", lock).Msg("lock released")
	}()
	// Second lock claim, this should refresh the lock and not err
	if _, err = capy.ClaimLock(lock); err != nil {
		log.Error().Err(err).Msg("unable to claim lock")
		return
	}

	log.Info().Str("lock", lock).Msg("lock is still ours")

	// Initialize another client that will try and steal the lock
	second, err := capybara.NewClient("127.0.0.1:8080", capybara.ClientOpts{
		Token:    "valid-token",
		CertPath: "certs/ca-cert.pem",
		Who:      "second",
	})
	if err != nil {
		log.Error().Err(err).Msg("unable to initialize capybara client")
		return
	}
	// Close
	defer func() {
		if err := second.Close(); err != nil {
			log.Error().Err(err).Msg("unable to close connection")
		}

		log.Info().Str("who", second.WhoAmI()).Msg("connection closed")
	}()

	// Attempt to claim lock
	if _, err = second.ClaimLock(lock); err != nil {
		log.Info().Str("lock", lock).Str("who", second.WhoAmI()).Err(err).Msg("unsuccessful attempt")
	}
}
