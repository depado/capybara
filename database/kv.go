package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	// ErrBucketNotFound is returned when a specific bucket can't be found.
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrNoBucket is returned when trying to put, get or delete a key with no
	// bucket.
	ErrNoBucket = errors.New("no bucket provided")
	// ErrIncompatibleValue is returned when attempting to put/delete/get a key
	// that is actually a bucket or a bucket that is actually a key. Basically
	// that means the bucket path + key is invalid.
	ErrIncompatibleValue = errors.New("incompatible value")
)

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

// Put puts a value at the given key in the given bucket. The buckets will be
// created on the fly if need be. An error will be returned if no bucket
// is provided or if the path is invalid.
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

	if errors.Is(err, bolt.ErrIncompatibleValue) {
		return ErrIncompatibleValue
	}

	return err
}

// PutPath puts a value at the given path. The buckets will be
// created on the fly if need be. An error will be returned if no bucket
// is provided or if the path is invalid.
func (cdb *CapybaraDB) PutPath(path, sep string, value []byte) error {
	o := strings.Split(path, sep)
	if len(o) < 2 {
		return ErrNoBucket
	}

	buckets, key := o[:len(o)-1], o[len(o)-1]

	return cdb.Put(buckets, key, value)
}

// Delete will attempt to delete the provided key in the given bucket path.
// An error is returned if the operation can't complete.
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

// DeletePath will delete a key given a full path to the key and a separator.
func (cdb *CapybaraDB) DeletePath(path, sep string) error {
	o := strings.Split(path, sep)
	if len(o) < 2 {
		return ErrNoBucket
	}

	buckets, key := o[:len(o)-1], o[len(o)-1]

	return cdb.Delete(buckets, key)
}

// Get returns the raw value of they key stored in the given bucket path.
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
		if b.Bucket([]byte(key)) != nil {
			return fmt.Errorf("key '%s' is a bucket: %w", key, ErrIncompatibleValue)
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

// GetPath will return the path.
func (cdb *CapybaraDB) GetPath(path, sep string) ([]byte, error) {
	o := strings.Split(path, sep)
	if len(o) < 2 {
		return nil, ErrNoBucket
	}

	buckets, key := o[:len(o)-1], o[len(o)-1]

	return cdb.Get(buckets, key)
}
