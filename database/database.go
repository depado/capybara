package database

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/fx"

	"github.com/Depado/capybara/cmd"
)

const (
	LocksBucket = "_locks"
)

var ErrLocksBucketNotFound = errors.New("locks bucket not found")

type CapybaraDB struct {
	db     *bolt.DB
	log    zerolog.Logger
	locksm sync.RWMutex
}

func NewCapybaraDB(lc fx.Lifecycle, conf *cmd.Conf, l zerolog.Logger) *CapybaraDB {
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
		log.Fatal().Err(err).Msg("unable to initialize buckets")
	}

	cdb := &CapybaraDB{
		db:  db,
		log: log,
	}

	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			cdb.log.Debug().Msg("closing database")
			return cdb.db.Close()
		},
	})

	return cdb
}
