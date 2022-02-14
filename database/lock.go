package database

import (
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Depado/capybara/pb"
)

var ErrNotOwner = errors.New("not the lock owner")
var ErrLockNotFound = errors.New("lock not found")

// ClaimLock can be used to claim a lock. If the lock is already owned, it will
// send back the lock's details. If the service creating this claim is the same
// as the owner (defined by the owner parameter), the lock's expiration date
// is delayed.
func (cdb *CapybaraDB) ClaimLock(key, owner string, pttl *time.Duration) (*pb.Lock, bool, error) {
	start := time.Now()
	var acquired bool

	var ttl time.Duration
	if pttl == nil {
		ttl = 5 * time.Minute
	} else {
		ttl = *pttl
	}

	cdb.locksm.Lock()
	defer cdb.locksm.Unlock()

	lock := &pb.Lock{}
	err := cdb.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(LocksBucket))
		if b == nil {
			return ErrLocksBucketNotFound
		}

		// Check for lock existence
		if raw := b.Get([]byte(key)); raw != nil {
			// The lock exists
			if err := proto.Unmarshal(raw, lock); err != nil {
				return fmt.Errorf("proto unmarshal: %w", err)
			}

			// Lock is not expired
			if lock.ValidUntil.AsTime().After(time.Now()) {
				if lock.Owner == owner {
					cdb.log.Debug().Str("lock", key).Str("owner", owner).Msg("lock is not expired but same owner, refresh")
					lock.ValidUntil = timestamppb.New(time.Now().Add(ttl))
					raw, err := proto.Marshal(lock)
					if err != nil {
						return fmt.Errorf("proto marshal: %w", err)
					}
					return b.Put([]byte(key), raw)
				}
				return nil
			}
			// Lock was expired, so we grant it
			cdb.log.Debug().Str("lock", key).Msg("lock was expired, claim granted")
		}

		// Insert lock
		acquired = true
		lock.Owner = owner
		lock.CreatedAt = timestamppb.Now()
		lock.ValidUntil = timestamppb.New(time.Now().Add(ttl))
		raw, err := proto.Marshal(lock)
		if err != nil {
			return fmt.Errorf("proto marshal: %w", err)
		}
		return b.Put([]byte(key), raw)
	})

	cdb.log.Debug().Str("took", time.Since(start).String()).Msg("lock claim completed")
	return lock, acquired, err
}

// ReleaseLock can be used to release (or free) a lock.
func (cdb *CapybaraDB) ReleaseLock(key, owner string) error {
	start := time.Now()

	cdb.locksm.Lock()
	defer cdb.locksm.Unlock()

	err := cdb.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(LocksBucket))
		if b == nil {
			return ErrLocksBucketNotFound
		}

		raw := b.Get([]byte(key))
		if raw == nil {
			return ErrLockNotFound
		}

		lock := &pb.Lock{}
		if err := proto.Unmarshal(raw, lock); err != nil {
			return fmt.Errorf("proto unmarshal: %w", err)
		}

		// Lock is already expired and shouldn't be in database
		if lock.ValidUntil.AsTime().Before(time.Now()) {
			if err := b.Delete([]byte(key)); err != nil {
				return fmt.Errorf("delete lock: %w", err)
			}
			return ErrLockNotFound
		}

		// Check ownership
		if lock.Owner != owner {
			return ErrNotOwner
		}

		// Actually delete the lock
		if err := b.Delete([]byte(key)); err != nil {
			return fmt.Errorf("delete lock: %w", err)
		}
		return nil
	})

	cdb.log.Debug().Str("took", time.Since(start).String()).Msg("lock release completed")
	return err
}
