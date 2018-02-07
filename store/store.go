package store

import (
	"errors"
	"time"
)

type Store interface {
	Get(key string) (val []byte, err error)
	Set(key string, val []byte, ttl time.Duration) error
	Update(key string, val []byte, ttl time.Duration) error
	Remove(key string) error
	Keys() []string
}

var (
	ErrNotFound            = errors.New("not found")
	ErrUnsuportedStoreType = errors.New("unsuported store type")
)
