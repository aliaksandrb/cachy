package server

import (
	"errors"
	"time"

	"github.com/aliaksandrb/cachy/server/mstore"
)

type Store interface {
	Get(key string) (val interface{}, found bool, err error)
	Set(key string, val interface{}, ttl time.Duration) error
	Remove(key string) error
}

type storeType int

const (
	MemoryStore = storeType(iota)
	PersistantStore
)

var (
	ErrInvalidStoreType = errors.New("invalid store type")
)

func New(st storeType) (Store, error) {
	if st == MemoryStore {
		// TODO pass at creation.
		return mstore.New(0)
	}

	return nil, ErrInvalidStoreType
}
