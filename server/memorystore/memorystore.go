package memorystore

import (
	"sync"
	"time"
)

var zeroTime = time.Time{}

func New() (*memorystore, error) {
	return &memorystore{
		s: make(map[string]*entry),
	}, nil
}

type memorystore struct {
	s  map[string]*entry
	mx sync.RWMutex
}

type entry struct {
	val interface{}
	ttl time.Time
}

func (e *entry) expired() bool {
	return e.ttl != zeroTime && time.Now().After(e.ttl)
}

func (m *memorystore) Get(key string) (val interface{}, found bool, err error) {
	m.mx.RLock()
	e, ok := m.s[key]
	m.mx.RUnlock()

	if !ok || e == nil || e.expired() {
		return nil, false, nil
	}

	return e.val, true, nil
}

func (m *memorystore) Set(key string, val interface{}, t time.Duration) error {
	var ttl time.Time
	if t == 0 {
		ttl = zeroTime
	} else {
		ttl = time.Now().Add(t)
	}

	m.mx.Lock()
	m.s[key] = &entry{val: val, ttl: ttl}
	m.mx.Unlock()

	return nil
}
