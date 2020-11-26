package server

import (
	"context"
	"errors"
	"time"

	"github.com/Depado/capybara/database"
	"github.com/Depado/capybara/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClaimLock implements the CapybaraServer interface.
// This function can be used to acquire a lock. If the lock is already owned
// by another owner, the function will return the lock's details such as its
// expiration date, creation date and who the owner is.
func (cap *CapybaraServer) ClaimLock(ctx context.Context, lr *pb.LockRequest) (*pb.LockResponse, error) {
	log := cap.log.With().Str("function", "ClaimLock").Logger()

	k := lr.GetKey()
	if k == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing key argument")
	}
	who := lr.GetWho()
	if who == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing who argument")
	}

	var ttl *time.Duration
	pbttl := lr.TTL
	if pbttl != nil {
		d := pbttl.AsDuration()
		ttl = &d
	}

	lock, ok, err := cap.db.ClaimLock(k, who, ttl)
	if err != nil {
		log.Err(err).Msg("unable to claim lock")
		return nil, status.Errorf(codes.Internal, "an error occured")
	}

	resp := &pb.LockResponse{
		Acquired:   ok,
		Owner:      lock.Owner,
		CreatedAt:  lock.CreatedAt,
		ValidUntil: lock.ValidUntil,
	}

	return resp, nil
}

func (cap *CapybaraServer) ReleaseLock(ctx context.Context, rr *pb.ReleaseRequest) (*pb.ReleaseResponse, error) {
	log := cap.log.With().Str("function", "ReleaseLock").Logger()

	k := rr.GetKey()
	if k == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing key argument")
	}
	who := rr.GetWho()
	if who == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing who argument")
	}

	err := cap.db.ReleaseLock(k, who)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrLockNotFound):
			return nil, status.Errorf(codes.NotFound, "lock not found")
		case errors.Is(err, database.ErrNotOwner):
			return nil, status.Errorf(codes.PermissionDenied, "not the owner of this lock")
		case errors.Is(err, database.ErrLocksBucketNotFound):
			log.Err(err).Msg("unable to release lock")
			return nil, status.Errorf(codes.Internal, "locks bucket can't be found")
		default:
			log.Err(err).Msg("unable to release lock")
			return nil, status.Errorf(codes.Internal, "unable to release lock")
		}
	}

	resp := &pb.ReleaseResponse{}
	return resp, nil
}
