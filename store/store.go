package store

import (
	"time"

	"github.com/aliaksandrb/cachy/store/mstore"
)

type Store interface {
	Get(key string) (val []byte, err error)
	Set(key string, val []byte, ttl time.Duration) error
	Update(key string, val []byte, ttl time.Duration) error
	Remove(key string) error
	Keys() []string
}

type Type int

const (
	MemoryStore = Type(iota)
	PersistantStore
)

func New(st Type) (Store, error) {
	if st == MemoryStore {
		// TODO pass at creation.
		return mstore.New(0)
	}

	return nil, ErrInvalidStoreType
}
