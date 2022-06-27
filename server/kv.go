package server

import (
	"context"
	"errors"
	"strings"

	"github.com/Depado/capybara/database"
	"github.com/Depado/capybara/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Put will insert data in the kv store.
func (cap *CapybaraServer) Put(ctx context.Context, pr *pb.PutRequest) (*pb.PutResponse, error) {
	if len(pr.Buckets) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one bucket required")
	}

	if pr.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key can't be empty")
	}

	if pr.Value == nil || len(pr.Value) == 0 {
		return nil, status.Error(codes.InvalidArgument, "value is nil or empty")
	}

	err := cap.db.Put(pr.Buckets, pr.Key, pr.Value)
	if err != nil {
		cap.log.Err(err).Str("buckets", strings.Join(pr.Buckets, "/")).Str("key", pr.Key).Msg("unable to put key")
		return nil, status.Error(codes.Internal, "unable to put key")
	}

	return &pb.PutResponse{}, nil
}

// Get will return data from the kv store (if any).
func (cap *CapybaraServer) Get(ctx context.Context, gr *pb.GetRequest) (*pb.GetResponse, error) {
	if len(gr.Buckets) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one bucket required")
	}

	if gr.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key can't be empty")
	}

	out, err := cap.db.Get(gr.Buckets, gr.Key)
	if err != nil {
		if errors.Is(err, database.ErrBucketNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		cap.log.Err(err).Str("buckets", strings.Join(gr.Buckets, "|")).Str("key", gr.Key).Msg("unable to get key")

		return nil, status.Error(codes.Internal, "unable to get key")
	}

	return &pb.GetResponse{Value: out}, nil
}

// Delete will delete data from the kv store.
func (cap *CapybaraServer) Delete(ctx context.Context, dr *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if len(dr.Buckets) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one bucket required")
	}

	if dr.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key can't be empty")
	}

	err := cap.db.Delete(dr.Buckets, dr.Key)
	if err != nil {
		if errors.Is(err, database.ErrBucketNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		cap.log.Err(err).Str("buckets", strings.Join(dr.Buckets, "|")).Str("key", dr.Key).Msg("unable to get key")

		return nil, status.Error(codes.Internal, "unable to get key")
	}

	return &pb.DeleteResponse{}, nil
}
