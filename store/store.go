package store

import (
	"errors"
	"time"
)

// Store is interface for backing storage.
type Store interface {
	// Get is to get a value by key. ErrNotFound if the key is missed.
	Get(key string) (val []byte, err error)
	// Set is to set a value for a key with ttl provided.
	Set(key string, val []byte, ttl time.Duration) error
	// Update is to update a value for a key provided. ErrNotFound if the key is missed.
	Update(key string, val []byte, ttl time.Duration) error
	// Remove removes a value by key.ErrNotFound if the key is missed.
	Remove(key string) error
	// Keys returns a slice of string keys existent in underlying store.
	// Currently it might return a staled keys too.
	Keys() []string
}

var (
	// ErrNotFound returned when there is not value for a key or it is expired.
	ErrNotFound = errors.New("not found")
	// ErrUnsuportedStoreType returned when store initialized with an unsuported type.
	ErrUnsuportedStoreType = errors.New("unsuported store type")
)
