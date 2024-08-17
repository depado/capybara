package database

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"

	"github.com/depado/capybara/cmd"
)

const (
	// LocksBucket is the default bucket used to store the locks.
	LocksBucket = "_locks"
)

// ErrLocksBucketNotFound is the error returned when the bucket isn't found.
var ErrLocksBucketNotFound = errors.New("locks bucket not found")

// CapybaraDB is the struct representing a capybara database.
type CapybaraDB struct {
	db     *bolt.DB
	log    zerolog.Logger
	locksm sync.RWMutex
}

// Close will close the database.
func (c *CapybaraDB) Close() error {
	c.log.Debug().Msg("closing database")
	return c.db.Close()
}

// NewCapybaraDB creates a new instance of CapybaraDB.
func NewCapybaraDB(conf *cmd.Conf, l zerolog.Logger) (*CapybaraDB, error) {
	log := l.With().Str("component", "database").Logger()

	db, err := bolt.Open(conf.Database.Path, 0666, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open database")
	}

	log.Debug().Msg("initialized")

	err = db.Update(func(t *bolt.Tx) error {
		_, err := t.CreateBucketIfNotExists([]byte(LocksBucket))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize buckets: %w", err)
	}

	cdb := &CapybaraDB{
		db:  db,
		log: log,
	}

	return cdb, nil
}
