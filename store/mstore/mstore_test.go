package mstore

import (
	"fmt"
	"sync/atomic"
	"testing"
)

/*
SHA: b1c49e9

$ go test -bench . -benchtime 5s -parallel 1 -cpu 4 ./store/mstore
goos: linux
goarch: amd64
pkg: github.com/aliaksandrb/cachy/store/mstore
BenchmarkSyncWrites-4          	30000000	       215 ns/op
BenchmarkConcurrentWrites-4    	20000000	       351 ns/op
BenchmarkSyncReads-4           	30000000	       261 ns/op
BenchmarkConcurrentReads-4     	30000000	       241 ns/op
BenchmarkSyncRemoves-4         	20000000	       491 ns/op ~276
BenchmarkConcurrentRemoves-4   	10000000	       701 ns/op ~350
PASS
ok  	github.com/aliaksandrb/cachy/store/mstore	47.748s

It is pretty rude. As purge lock is not counted.
*/

var testKey = "836fa3665997a860728bcb9e9a1e704d427cfc920e79d847d79c8a9a907b9e965defa4154b2b86bdec6930adbe33f21364523a6f6ce363865724549fdfc08553"
var testVal = []byte(testKey)

func BenchmarkSyncWrites(b *testing.B) {
	var err error
	store, _ := New(1, 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err = store.Set(testKey, testVal, 0); err != nil {
			b.Fatalf("unexpected error during benchmark: %v", err)
		}
	}
}

func BenchmarkConcurrentWrites(b *testing.B) {
	var err error
	store, _ := New(1, 1)

	b.SetParallelism(100)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err = store.Set(testKey, testVal, 0); err != nil {
				b.Fatalf("unexpected error during benchmark: %v", err)
			}
		}
	})
}

func BenchmarkSyncReads(b *testing.B) {
	var err error
	var val []byte

	store, _ := New(1, 1)
	if err = store.Set(testKey, testVal, 0); err != nil {
		b.Fatalf("unexpected error during benchmark: %v", err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if val, err = store.Get(testKey); err != nil {
			b.Fatalf("unexpected error during benchmark: %v - %q", err, val)
		}
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	var err error
	var val []byte

	store, _ := New(1, 1)
	if err = store.Set(testKey, testVal, 0); err != nil {
		b.Fatalf("unexpected error during benchmark: %v", err)
	}

	b.SetParallelism(100)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if val, err = store.Get(testKey); err != nil {
				b.Fatalf("unexpected error during benchmark: %v - %q", err, val)
			}
		}
	})
}

func BenchmarkSyncRemoves(b *testing.B) {
	var err error

	store, _ := New(1, 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err = store.Set(testKey, testVal, 0); err != nil {
			b.Fatalf("unexpected error during benchmark: %v", err)
		}

		if err = store.Remove(testKey); err != nil {
			b.Fatalf("unexpected error during benchmark: %v", err)
		}
	}
}

func BenchmarkConcurrentRemoves(b *testing.B) {
	var err error
	var i int32

	store, _ := New(1, 1)
	b.SetParallelism(100)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		num := atomic.AddInt32(&i, 1)
		key := fmt.Sprintf("%d", num)

		for pb.Next() {
			if err = store.Set(key, testVal, 0); err != nil {
				b.Fatalf("unexpected error during benchmark: %v", err)
			}

			if err = store.Remove(key); err != nil {
				b.Fatalf("unexpected error during benchmark: %v - %v", err, key)
			}
		}
	})
}
