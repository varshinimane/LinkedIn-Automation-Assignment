package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketMeta     = []byte("meta")
	bucketCounters = []byte("counters")
)

// Store wraps BoltDB for simple structured storage.
type Store struct {
	db *bolt.DB
}

type Metadata struct {
	LastLoginAt time.Time `json:"last_login_at"`
}

// Open initializes the DB and buckets.
func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open state db: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(bucketMeta); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(bucketCounters); e != nil {
			return e
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close flushes the DB.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// SaveMetadata persists run-level metadata.
func (s *Store) SaveMetadata(meta Metadata) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMeta)
		raw, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		return b.Put([]byte("meta"), raw)
	})
}

// LoadMetadata retrieves stored metadata.
func (s *Store) LoadMetadata() (Metadata, error) {
	var meta Metadata
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMeta)
		raw := b.Get([]byte("meta"))
		if len(raw) == 0 {
			return errors.New("not found")
		}
		return json.Unmarshal(raw, &meta)
	})
	return meta, err
}

// IncrementCounter increments the named counter atomically and returns the new value.
func (s *Store) IncrementCounter(key string) (int, error) {
	var newVal int
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCounters)
		current := 0
		raw := b.Get([]byte(key))
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &current); err != nil {
				return err
			}
		}
		current++
		enc, err := json.Marshal(current)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(key), enc); err != nil {
			return err
		}
		newVal = current
		return nil
	})
	return newVal, err
}

// ResetCounter sets the counter to zero.
func (s *Store) ResetCounter(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCounters)
		enc, _ := json.Marshal(0)
		return b.Put([]byte(key), enc)
	})
}

// GetCounter reads the current value of a counter.
func (s *Store) GetCounter(key string) (int, error) {
	val := 0
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCounters)
		raw := b.Get([]byte(key))
		if len(raw) == 0 {
			val = 0
			return nil
		}
		return json.Unmarshal(raw, &val)
	})
	return val, err
}

