package server

import (
	"errors"
	"time"

	"github.com/aliaksandrb/cachy/server/memorystore"
)

type Store interface {
	Get(key string) (val interface{}, found bool, err error)
	Set(key string, val interface{}, ttl time.Duration) error
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
		return memorystore.New()
	}

	return nil, ErrInvalidStoreType
}
