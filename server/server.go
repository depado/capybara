package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/depado/capybara/cmd"
	"github.com/depado/capybara/database"
	"github.com/depado/capybara/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// CapybaraServer represents the GRPC server.
type CapybaraServer struct {
	db  *database.CapybaraDB
	log zerolog.Logger
	pb.UnimplementedCapybaraServer
}

// NewGRPCServer will create a new GRPC server given the proper configuration,
// logger and database config.
func NewGRPCServer(conf *cmd.Conf, l zerolog.Logger, cdb *database.CapybaraDB) (*grpc.Server, error) {
	var gs *grpc.Server

	cap := &CapybaraServer{
		db:  cdb,
		log: l.With().Str("component", "grpc").Logger(),
	}

	if conf.Server.TLS.CertPath != "" && conf.Server.TLS.KeyPath != "" {
		tlsCredentials, err := loadTLSCredentials(conf.Server.TLS.CertPath, conf.Server.TLS.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("load TLS credentials: %w", err)
		}

		l.Info().Str("cert", conf.Server.TLS.CertPath).Str("key", conf.Server.TLS.KeyPath).Msg("loaded credentials")

		gs = grpc.NewServer(grpc.Creds(tlsCredentials), grpc.ChainUnaryInterceptor(cap.AuthInterceptor))
	} else {
		gs = grpc.NewServer(grpc.ChainUnaryInterceptor(cap.AuthInterceptor))
	}

	pb.RegisterCapybaraServer(gs, cap)

	return gs, nil
}

// Listen will start the GRPC server and listen on the configured port/host.
func Listen(conf *cmd.Conf, log zerolog.Logger, gs *grpc.Server) {
	la := conf.Server.ListenAddr()

	lis, err := net.Listen("tcp", la)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	log.Info().Str("address", la).Msg("listening")

	if err := gs.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func loadTLSCredentials(cert, key string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
