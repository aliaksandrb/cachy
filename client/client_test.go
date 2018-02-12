package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/aliaksandrb/cachy/server"
	"github.com/aliaksandrb/cachy/store"
)

func TestClient(t *testing.T) {
	skipShort(t)
	time.Sleep(50 * time.Millisecond)

	server, err := server.Run(server.MemoryStore, 5, ":3000")
	checkErr(t, err)
	defer server.Stop()

	session, err := New("127.0.0.1:3000", 3)
	checkErr(t, err)
	defer session.Close()

	for i, tc := range []struct {
		key string
		in  interface{}
	}{
		{
			key: "key",
			in:  "string",
		}, {
			key: "nil",
			in:  interface{}(nil),
		}, {
			key: "map",
			in:  map[interface{}]interface{}{"hello": "world", "cool": 1},
		}, {
			key: "slice",
			in:  []interface{}{"hello", "world"},
		}, {
			key: "int",
			in:  42,
		},
	} {

		if err = session.Set(tc.key, tc.in, 0); err != nil {
			t.Errorf("[%d] unable to store new value: %+v, err: %v", i, tc, err)
			continue
		}

		got, err := session.Get(tc.key)
		if err != nil {
			t.Errorf("[%d] unable to get a value: %q, err: %v", i, tc.key, err)
			continue
		}

		if !reflect.DeepEqual(got, tc.in) {
			t.Errorf("[%d] should be equal: got %q, want %q", i, got, tc.in)
		}
	}
}

func TestClientExpiration(t *testing.T) {
	skipShort(t)
	time.Sleep(50 * time.Millisecond)

	server, err := server.Run(server.MemoryStore, 5, ":3000")
	checkErr(t, err)
	defer server.Stop()

	session, err := New("127.0.0.1:3000", 3)
	checkErr(t, err)
	defer session.Close()

	want := map[interface{}]interface{}{"some": 1, "another": []interface{}{1, 2, 3}}
	key := "key"

	if err = session.Set(key, want, 50*time.Millisecond); err != nil {
		t.Fatalf("unable to store new value: %+v, err: %v", want, err)
	}

	got, err := session.Get(key)
	if err != nil {
		t.Fatalf("unable to get a value: %q, err: %v", key, err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("should be equal: got %q, want %q", got, want)
	}

	time.Sleep(50 * time.Millisecond)

	got, err = session.Get(key)
	if err == nil || !reflect.DeepEqual(err, store.ErrNotFound) {
		t.Fatalf("should be error: got: %v, want: %v, err: %v", got, store.ErrNotFound, err)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func skipShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping long running tests in a -short mode")
	}
}
