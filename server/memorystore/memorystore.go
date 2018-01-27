package memorystore

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var zeroTime = time.Time{}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func New() (*memorystore, error) {
	ms := &memorystore{
		s:    make(map[string]*entry),
		keys: make([]string, 1),
	}
	go ms.startPurger()

	return ms, nil
}

type memorystore struct {
	s      map[string]*entry
	purger *purger
	keys   []string

	mx sync.RWMutex
}

func (m *memorystore) startPurger() {
	m.purger = &purger{
		purgeQueue: make(chan string),
		quit:       make(chan struct{}),
		store:      m,
	}
	m.purger.Start()
}

func (m *memorystore) Remove(key string) error {
	// It does lock reads too
	m.mx.Lock()
	delete(m.s, key)
	m.mx.Unlock()
	return nil
}

func (m *memorystore) Get(key string) (val interface{}, found bool, err error) {
	m.mx.RLock()
	e, ok := m.s[key]
	m.mx.RUnlock()

	if ok && e != nil && !e.expired() {
		return e.val, true, nil
	}

	if ok {
		go func() { m.purger.purgeQueue <- key }()
	}

	return nil, false, nil
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
	m.keys = append(m.keys, key)
	m.mx.Unlock()

	return nil
}

func (m *memorystore) RefreshKeys() {
	m.mx.RLock()
	m.keys = m.keys[:0]
	for k, _ := range m.s {
		m.keys = append(m.keys, k)
	}
	m.mx.RUnlock()
}

type purger struct {
	purgeQueue chan string
	quit       chan struct{}
	store      *memorystore
}

func (p *purger) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

Loop:
	for {
		select {
		case key := <-p.purgeQueue:
			if err := p.store.Remove(key); err != nil {
				fmt.Println("key %v is already evicted: %v", key, err)
			}
		case <-ticker.C:
			p.purgeStaleKeys()
		case <-p.quit:
			break Loop
		}
	}
}

func (p *purger) purgeStaleKeys() {
	var key string
	size := len(p.store.keys)
	quater := size / 4
	if quater < 2 {
		quater = size
	}

	for i := 0; i < quater; i++ {
		key = p.store.keys[rand.Intn(size)]
		if _, _, err := p.store.Get(key); err != nil {
			fmt.Println("key %v is already evicted: %v", key, err)
		}
	}

	p.store.RefreshKeys()
}

func (p *purger) Stop() {
	close(p.quit)
}

type entry struct {
	val interface{}
	ttl time.Time
}

func (e *entry) expired() bool {
	return e.ttl != zeroTime && time.Now().After(e.ttl)
}
