package store

import (
	"errors"
	"time"

	"github.com/aliaksandrb/cachy/store/mstore"
)

type Store interface {
	Get(key string) (val interface{}, found bool, err error)
	Set(key string, val interface{}, ttl time.Duration) error

	Update(key string, val interface{}, ttl time.Duration) error
	Remove(key string) error
	Keys() []string
}

type Type int

const (
	MemoryStore = Type(iota)
	PersistantStore
)

var (
	ErrInvalidStoreType = errors.New("invalid store type")
)

func New(st Type) (Store, error) {
	if st == MemoryStore {
		// TODO pass at creation.
		return mstore.New(0)
	}

	return nil, ErrInvalidStoreType
}
