package mstore

import (
	"sync"
	"time"

	"github.com/aliaksandrb/cachy/store"
	"github.com/spaolacci/murmur3"
)

const defaultBucketsNum int = 3
const defaultPurgeInterval int = 10

var zeroTime = time.Time{}

func New(bucketsNum int, purgeInterval int) (*mStore, error) {
	if bucketsNum == 0 {
		bucketsNum = defaultBucketsNum
	}
	if purgeInterval == 0 {
		purgeInterval = defaultPurgeInterval
	}

	buckets := make([]*bucket, bucketsNum)
	for i := range buckets {
		buckets[i] = newBucket()
	}

	m := &mStore{
		purgeInterval: purgeInterval,
		buckets:       buckets,
		purger:        newPurger(),
	}

	go m.startPurger()

	return m, nil
}

// mStore implements store.Store
type mStore struct {
	purgeInterval int
	buckets       []*bucket
	purger        *purger
}

func (m *mStore) bucketsNum() int {
	// TODO is fine for fixed-size buckets
	return len(m.buckets)
}

func (m *mStore) getBucket(key string) *bucket {
	return m.buckets[m.hash([]byte(key))]
}

func (m *mStore) hash(k []byte) uint64 {
	return murmur3.Sum64(k) % uint64(m.bucketsNum())
}

type bucket struct {
	s map[string]*entry

	mu sync.RWMutex
}

func newBucket() *bucket {
	return &bucket{
		s: make(map[string]*entry),
	}
}

func (m *mStore) startPurger() {
	ticker := time.NewTicker(time.Duration(m.purgeInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.purger.quit:
			return
		case <-ticker.C:
			for _, b := range m.buckets {
				m.purger.purgeStaleKeys(b)
			}
		}
	}
}

func (m *mStore) Get(key string) (val []byte, err error) {
	b := m.getBucket(key)
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.s[key]

	if !ok {
		return nil, store.ErrNotFound
	}

	// TODO entry level lock rather than bucket?
	if e == nil || e.expired() {
		go m.Remove(key)
	}

	return append([]byte(nil), e.val...), nil
}

func getTTL(t time.Duration) time.Time {
	if t == 0 {
		return zeroTime
	}

	return time.Now().Add(t)
}

func (m *mStore) Set(key string, val []byte, t time.Duration) error {
	b := m.getBucket(key)
	b.mu.Lock()
	b.s[key] = &entry{val: val, ttl: getTTL(t)}
	b.mu.Unlock()

	return nil
}

func (m *mStore) Update(key string, val []byte, t time.Duration) error {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.s[key]
	if !ok {
		return store.ErrNotFound
	}

	e.val = val
	e.ttl = getTTL(t)
	b.s[key] = e

	return nil
}

func (m *mStore) Remove(key string) error {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.s[key]
	if !ok {
		return store.ErrNotFound
	}
	delete(b.s, key)

	return nil
}

func (m *mStore) Keys() (keys []string) {
	for _, b := range m.buckets {
		b.mu.RLock()
		for k := range b.s {
			keys = append(keys, k)
		}
		b.mu.RUnlock()
	}

	return
}

type purger struct {
	quit chan struct{}
}

func newPurger() *purger {
	return &purger{quit: make(chan struct{})}
}

func (p *purger) purgeStaleKeys(b *bucket) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// TODO make lock shorter

	for k, e := range b.s {
		if e == nil || e.expired() {
			delete(b.s, k)
		}
	}
}

type entry struct {
	val []byte
	ttl time.Time
}

func (e *entry) expired() bool {
	return e.ttl != zeroTime && time.Now().After(e.ttl)
}
