package mstore

import (
	"fmt"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"

	log "github.com/aliaksandrb/cachy/logger"
)

// TODO the more the better for concurent access.
const defaultBucketsNum int = 3

var zeroTime = time.Time{}

func New(bucketsNum int) (*mStore, error) {
	if bucketsNum == 0 {
		bucketsNum = defaultBucketsNum
	}

	buckets := make([]*bucket, bucketsNum)
	for i := range buckets {
		buckets[i] = newBucket()
	}

	m := &mStore{
		buckets: buckets,
		purger:  newPurger(),
	}

	go m.startPurger()

	return m, nil
}

// mStore implements store.Store
type mStore struct {
	buckets []*bucket
	purger  *purger
}

func (m *mStore) String() string {
	return fmt.Sprintf(`
		mStore {
			bucketsNum: %d
			keys: %v
			buckets: {
				%s
			}
		}
	`, m.bucketsNum(), m.Keys(), m.buckets)
}

func (m *mStore) bucketsNum() int {
	// TODO is fine for fixed-size buckets
	return len(m.buckets)
}

// TODO keys as ints
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

func (b *bucket) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return fmt.Sprintf(`
		%+v
	`, b.s)
}

func newBucket() *bucket {
	return &bucket{
		s: make(map[string]*entry),
	}
}

func (m *mStore) startPurger() {
	ticker := time.NewTicker(5 * time.Second)
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

func (m *mStore) Get(key string) (val interface{}, found bool, err error) {
	b := m.getBucket(key)
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.s[key]

	if !ok {
		return nil, false, nil
	}

	// TODO entry level lock rather than bucket?
	if e == nil || e.expired() {
		go m.Remove(key)
	}

	return e.val, true, nil
}

func getTTL(t time.Duration) time.Time {
	if t == 0 {
		return zeroTime
	}

	return time.Now().Add(t)
}

func (m *mStore) Set(key string, val interface{}, t time.Duration) error {
	log.Info("set %s %+v %v", key, val, t)
	b := m.getBucket(key)
	b.mu.Lock()
	b.s[key] = &entry{val: val, ttl: getTTL(t)}
	b.mu.Unlock()

	return nil
}

func (m *mStore) Update(key string, val interface{}, t time.Duration) error {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.s[key]
	if !ok {
		return nil
	}

	e.val = val
	e.ttl = getTTL(t)
	b.s[key] = e

	return nil
}

func (m *mStore) Remove(key string) error {
	b := m.getBucket(key)
	b.mu.Lock()
	delete(b.s, key)
	b.mu.Unlock()
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
	// TODO make lock shorter

	for k, e := range b.s {
		if e == nil || e.expired() {
			delete(b.s, k)
		}
	}
	//var key string
	//size := len(p.store.keys)
	//quater := size / 4
	//if quater < 2 {
	//	quater = size
	//}

	//for i := 0; i < quater; i++ {
	//	key = p.store.keys[rand.Intn(size)]
	//	if _, _, err := p.store.Get(key); err != nil {
	//		fmt.Println("key %v is already evicted: %v", key, err)
	//	}
	//}

	b.mu.Unlock()
}

type entry struct {
	val interface{}
	ttl time.Time
}

func (e *entry) expired() bool {
	return e.ttl != zeroTime && time.Now().After(e.ttl)
}

func (e *entry) String() string {
	return fmt.Sprintf(`
		{
			val: %v
			ttl: %v
		}
	`, e.val, e.ttl)
}
