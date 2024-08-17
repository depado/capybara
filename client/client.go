package capybara

import (
	"context"
	"errors"
	"fmt"

	"github.com/depado/capybara/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ErrLockNotClaimed is an error returned when the desired lock can't be
// claimed. This happens if the lock is already claimed by another client.
var ErrLockNotClaimed = errors.New("lock not claimed")

// Client is the main client struct. Use NewClient to initialize a new one.
type Client struct {
	who  string
	ctx  context.Context
	capy pb.CapybaraClient
	conn *grpc.ClientConn
}

// ClientOpts represents the various options that can be passed to NewClient.
// It allows to customize the behavior of the client.
//
// Token: A token to enforce security between the client and the server. This
// token will be passed to all subsequent requests made with the client.
// CertPath: Path to the server's certificate authority certiticate. Given a
// proper certificate, this option will ensure the connection is encrypted.
// Who: Unique identifier. If this option isn't provided, a unique ID will be
// generated on the fly.
type ClientOpts struct {
	Token    string
	CertPath string
	Who      string
}

// NewClient creates a new capybara client using the given capybara GRPc address
// and various options such as the token, path to the server's CA certificate
// or a unique identifier.
func NewClient(addr string, opts ClientOpts) (*Client, error) {
	var (
		err error
		ctx context.Context
		who string
	)

	creds := insecure.NewCredentials()

	if opts.Token != "" {
		ctx = metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{"token": opts.Token}),
		)
	}

	if opts.CertPath != "" {
		if creds, err = loadTLSCredentials(opts.CertPath); err != nil {
			return nil, fmt.Errorf("load credentials: %w", err)
		}
	}

	if opts.Who != "" {
		who = opts.Who
	} else {
		who = uuid.NewString()
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return &Client{who: who, ctx: ctx, capy: pb.NewCapybaraClient(conn), conn: conn}, nil
}

// Close will close the internal grpc connection.
func (c Client) Close() error {
	return c.conn.Close()
}

// WhoAmI returns the unique identifier associated with the client. It can
// reflect the one that was provided when initializing the client or the one
// generated if no unique identifier was provided.
func (c Client) WhoAmI() string {
	return c.who
}

// ClaimLockRaw claims the given lock and returns the LockResponse.
// The pb.LockResponse.Acquired field should be checked to confirm the lock
// has been acquired.
func (c Client) ClaimLockRaw(lock string) (*pb.LockResponse, error) {
	pr, err := c.capy.ClaimLock(c.ctx, &pb.LockRequest{Key: lock, Who: c.who})
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// ClaimLock will claim the given lock and return an error if it can't
// successfully claim said lock.
func (c Client) ClaimLock(lock string) (*pb.LockResponse, error) {
	pr, err := c.capy.ClaimLock(c.ctx, &pb.LockRequest{Key: lock, Who: c.who})
	if err != nil {
		return nil, err
	}

	if !pr.Acquired && pr.Owner != c.who {
		return pr, ErrLockNotClaimed
	}

	return pr, nil
}

func (c Client) ReleaseLock(lock string) error {
	_, err := c.capy.ReleaseLock(c.ctx, &pb.ReleaseRequest{Key: lock, Who: c.who})
	return err
}
