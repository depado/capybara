package server

import (
	"net"

	"github.com/Depado/capybara/cmd"
	"github.com/Depado/capybara/database"
	"github.com/Depado/capybara/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type CapybaraServer struct {
	db  *database.CapybaraDB
	log zerolog.Logger
	pb.UnimplementedCapybaraServer
}

func NewGRPCServer(conf *cmd.Conf, l zerolog.Logger, cdb *database.CapybaraDB) *grpc.Server {
	cap := &CapybaraServer{
		db:  cdb,
		log: l.With().Str("component", "grpc").Logger(),
	}

	gs := grpc.NewServer(grpc.ChainUnaryInterceptor(cap.AuthInterceptor))
	pb.RegisterCapybaraServer(gs, cap)
	return gs
}

func Listen(conf *cmd.Conf, log zerolog.Logger, gs *grpc.Server) error {
	la := conf.Server.ListenAddr()
	lis, err := net.Listen("tcp", la)
	if err != nil {
		return err
	}

	go func() {
		if err := gs.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to serve")
		}
	}()
	log.Info().Str("address", la).Msg("listening")
	return nil
}
