package main

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/depado/capybara/pb"
)

func main() {
	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Connect to capybara server
	conn, err := grpc.NewClient("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open grpc connection")
	}
	defer conn.Close() //nolint:errcheck

	// Initialize capybara client and set context with a valid token
	lc := pb.NewCapybaraClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"token": "valid-token"}))
	whoami := uuid.New().String()
	lock := "hello"

	// Claim a lock and log timings
	for i := 0; i < 3; i++ {
		start := time.Now()

		pr, err := lc.ClaimLock(ctx, &pb.LockRequest{Key: lock, Who: whoami})
		if err != nil {
			log.Fatal().Err(err).Msg("unable to call claimlock")
		}

		log.Info().Bool("owned", pr.Owner == whoami).
			Bool("acquired", pr.Acquired).Str("owner", pr.Owner).
			Time("created_at", pr.CreatedAt.AsTime()).
			Time("valid_until", pr.ValidUntil.AsTime()).
			Str("took", time.Since(start).String()).
			Msg("lock result")

		time.Sleep(100 * time.Millisecond)
	}

	// Release the claimed lock
	start := time.Now()

	if _, err = lc.ReleaseLock(ctx, &pb.ReleaseRequest{Key: lock, Who: whoami}); err != nil {
		log.Fatal().Err(err).Msg("unable to release lock")
	}

	log.Info().Str("lock", lock).Str("whoami", whoami).Str("took", time.Since(start).String()).Msg("released lock")

	// Release the claimed lock again (failing on purpose)
	start = time.Now()

	if _, err = lc.ReleaseLock(ctx, &pb.ReleaseRequest{Key: lock, Who: whoami}); err != nil {
		log.Info().Str("lock", lock).Str("whoami", whoami).
			Str("took", time.Since(start).String()).Msg("lock already released")
	}

	content := []byte("noconf")
	buckets := []string{"guilds", "12345678"}
	key := "conf"

	start = time.Now()

	_, err = lc.Put(ctx, &pb.PutRequest{
		Buckets: buckets,
		Value:   content,
		Key:     key,
	})
	if err != nil {
		log.Err(err).Msg("unable to put key")
	}

	log.Info().Strs("buckets", buckets).Str("key", key).
		Str("value", string(content)).Str("took", time.Since(start).String()).
		Msg("successful put")

	start = time.Now()

	out, err := lc.Get(ctx, &pb.GetRequest{Buckets: buckets, Key: key})
	if err != nil {
		log.Err(err).Msg("unable to retrieve key")
	} else {
		log.Info().Strs("buckets", buckets).Str("key", key).
			Str("value", string(out.Value)).
			Str("took", time.Since(start).String()).Msg("successful get")
	}

	start = time.Now()

	_, err = lc.Delete(ctx, &pb.DeleteRequest{Buckets: buckets, Key: key})
	if err != nil {
		log.Err(err).Msg("unable to delete item")
	} else {
		log.Info().Strs("buckets", buckets).Str("key", key).
			Str("value", string(content)).Str("took", time.Since(start).String()).
			Msg("successful del")
	}

	start = time.Now()

	out, err = lc.Get(ctx, &pb.GetRequest{Buckets: buckets, Key: key})
	if err != nil {
		log.Info().Strs("buckets", buckets).Str("key", key).
			Str("value", string(out.Value)).Str("took", time.Since(start).String()).
			Err(err).Msg("failed get")
	}

	start = time.Now()

	_, err = lc.Put(ctx, &pb.PutRequest{
		// bucket/nested/firstbucket - firstkey: test
		Buckets: []string{"bucket", "nested", "firstbucket"},
		Key:     "firstkey",
		Value:   []byte("test"),
	})
	if err != nil {
		log.Err(err).Msg("unable to put key")
	}

	log.Info().Strs("buckets", buckets).Str("key", key).
		Str("value", string(content)).Str("took", time.Since(start).String()).
		Msg("successful put")

	start = time.Now()
	buckets = []string{"bucket", "nested", "secondbucket"}

	if _, err = lc.Put(ctx, &pb.PutRequest{
		// bucket/nested/secondbucket - firstkey: test
		Buckets: buckets,
		Key:     "firstkey",
		Value:   []byte("test"),
	}); err != nil {
		log.Err(err).Msg("unable to put key")
	}

	log.Info().Strs("buckets", buckets).Str("key", "firstkey").
		Str("value", string(content)).Str("took", time.Since(start).String()).
		Msg("successful put")

	start = time.Now()

	if _, err = lc.Put(ctx, &pb.PutRequest{
		Buckets: []string{"bucket", "nested"},
		Value:   content,
		Key:     "key",
	}); err != nil {
		log.Err(err).Msg("unable to put key")
	}

	log.Info().Strs("buckets", buckets).Str("key", key).
		Str("value", string(content)).Str("took", time.Since(start).String()).
		Msg("successful put")

	start = time.Now()

	out, err = lc.Get(ctx, &pb.GetRequest{Buckets: []string{"bucket", "nested"}, Key: "key"})
	if err != nil {
		log.Err(err).Msg("unable to retrieve key")
	} else {
		log.Info().Strs("buckets", buckets).Str("key", key).
			Str("value", string(out.Value)).
			Str("took", time.Since(start).String()).Msg("successful get")
	}
}
