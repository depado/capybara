package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor intercepts incoming grpc calls and will fetch the
// authentication token in the context.
// TODO: True check
func (cap *CapybaraServer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Debug().Msg("unauthenticated request")
		return nil, status.Errorf(codes.Unauthenticated, "missing context metadata")
	}

	if len(meta["token"]) != 1 {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	if meta["token"][0] != "valid-token" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return handler(ctx, req)
}
