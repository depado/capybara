package database

import (
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrBucketNotFound = errors.New("bucket not found")
var ErrNoBucket = errors.New("no bucket provided")

// TraverseCreate will traverse the whole bucket tree defined in the buckets
// argument and will create the necessary buckets if they don't exist.
// This function will return the last bucket.
func TraverseCreate(t *bolt.Tx, buckets []string) (*bolt.Bucket, error) {
	b, err := t.CreateBucketIfNotExists([]byte(buckets[0]))
	if err != nil {
		return nil, err
	}
	for _, bk := range buckets[1:] {
		if b, err = b.CreateBucketIfNotExists([]byte(bk)); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Traverse will traverse the whole bucket tree defined in the buckets argument
// and will fail if a bucket isn't found. This function should be used to find
// the appropriate bucket for delete and get operations (since they should
// always be existing keys).
func Traverse(t *bolt.Tx, buckets []string) (*bolt.Bucket, error) {
	b := t.Bucket([]byte(buckets[0]))
	if b == nil {
		return nil, fmt.Errorf("bucket %s: %w", buckets[0], ErrBucketNotFound)
	}
	for _, bk := range buckets[1:] {
		if b = b.Bucket([]byte(bk)); b == nil {
			return nil, fmt.Errorf("bucket %s: %w", bk, ErrBucketNotFound)
		}
	}
	return b, nil
}

func (cdb *CapybaraDB) Put(buckets []string, key string, value []byte) error {
	start := time.Now()
	defer cdb.log.Debug().Str("took", time.Since(start).String()).Str("key", key).Str("action", "put").Send()
	if len(buckets) == 0 {
		return ErrNoBucket
	}

	err := cdb.db.Update(func(t *bolt.Tx) error {
		b, err := TraverseCreate(t, buckets)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})

	return err
}

func (cdb *CapybaraDB) Delete(buckets []string, key string) error {
	start := time.Now()
	defer cdb.log.Debug().Str("took", time.Since(start).String()).Str("key", key).Str("action", "delete").Send()
	if len(buckets) == 0 {
		return ErrNoBucket
	}

	err := cdb.db.Update(func(t *bolt.Tx) error {
		b, err := Traverse(t, buckets)
		if err != nil {
			return err
		}
		return b.Delete([]byte(key))
	})

	return err
}

func (cdb *CapybaraDB) Get(buckets []string, key string) ([]byte, error) {
	start := time.Now()
	defer cdb.log.Debug().Str("took", time.Since(start).String()).Str("key", key).Str("action", "get").Send()
	if len(buckets) == 0 {
		return nil, ErrNoBucket
	}

	var out []byte

	err := cdb.db.View(func(t *bolt.Tx) error {
		b, err := Traverse(t, buckets)
		if err != nil {
			return err
		}
		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}
		out = make([]byte, len(v))
		copy(out, v)
		return nil
	})

	return out, err
}
