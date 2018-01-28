package mstore

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

// TODO the more the better for concurent access.
const defaultBucketsNum int = 32

var zeroTime = time.Time{}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TODO to be an interface Store
func New(bucketsNum int) (*mStore, error) {
	if bucketsNum == 0 {
		bucketsNum = defaultBucketsNum
	}

	buckets := make([]*bucket, bucketsNum)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	//go m.startPurger()

	return &mStore{
		buckets: buckets,
	}, nil
}

type mStore struct {
	buckets []*bucket
}

func (m *mStore) String() string {
	return fmt.Sprintf(`
		mStore {
			bucketsNum: %d
			buckets: {
				%s
			}
		}
	`, m.bucketsNum(), m.buckets)
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
	//purger *purger
	//keys   []string

	mu sync.RWMutex
}

func (b *bucket) String() string {
	return fmt.Sprintf(`
		%+v
	`, b.s)
}

func newBucket() *bucket {
	return &bucket{
		s: make(map[string]*entry),
		//keys: make([]string, 1),
	}
}

//func (m *mStore) startPurger() {
//	m.purger = &purger{
//		purgeQueue: make(chan string),
//		quit:       make(chan struct{}),
//		store:      m,
//	}
//	m.purger.Start()
//}

func (m *mStore) Remove(key string) error {
	b := m.getBucket(key)
	b.mu.Lock()
	delete(b.s, key)
	b.mu.Unlock()
	return nil
}

func (m *mStore) Get(key string) (val interface{}, found bool, err error) {
	b := m.getBucket(key)
	b.mu.RLock()
	e, ok := b.s[key]
	b.mu.RUnlock()

	if ok && e != nil && !e.expired() {
		return e.val, true, nil
	}

	///	if ok {
	///		go func() { m.purger.purgeQueue <- key }()
	///	}

	return nil, false, nil
}

func (m *mStore) Set(key string, val interface{}, t time.Duration) error {
	var ttl time.Time
	if t == 0 {
		ttl = zeroTime
	} else {
		ttl = time.Now().Add(t)
	}

	b := m.getBucket(key)
	b.mu.Lock()
	b.s[key] = &entry{val: val, ttl: ttl}
	//b.keys = append(m.keys, key)
	b.mu.Unlock()

	return nil
}

//func (m *memorystore) RefreshKeys() {
//	m.mx.RLock()
//	m.keys = m.keys[:0]
//	for k, _ := range m.s {
//		m.keys = append(m.keys, k)
//	}
//	m.mx.RUnlock()
//}

//type purger struct {
//	purgeQueue chan string
//	quit       chan struct{}
//	store      *memorystore
//}
//
//func (p *purger) Start() {
//	ticker := time.NewTicker(5 * time.Second)
//	defer ticker.Stop()
//
//Loop:
//	for {
//		select {
//		case key := <-p.purgeQueue:
//			if err := p.store.Remove(key); err != nil {
//				fmt.Println("key %v is already evicted: %v", key, err)
//			}
//		case <-ticker.C:
//			p.purgeStaleKeys()
//		case <-p.quit:
//			break Loop
//		}
//	}
//}
//
//func (p *purger) purgeStaleKeys() {
//	var key string
//	size := len(p.store.keys)
//	quater := size / 4
//	if quater < 2 {
//		quater = size
//	}
//
//	for i := 0; i < quater; i++ {
//		key = p.store.keys[rand.Intn(size)]
//		if _, _, err := p.store.Get(key); err != nil {
//			fmt.Println("key %v is already evicted: %v", key, err)
//		}
//	}
//
//	p.store.RefreshKeys()
//}
//
//func (p *purger) Stop() {
//	close(p.quit)
//}

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
