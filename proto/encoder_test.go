package proto

import (
	"bytes"
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestEncode(t *testing.T) {
	var (
		nullInterface interface{}
		nullErr       error
		nullString    string
		nullInt       int
		nullSlice     []interface{}
		nullMap       map[interface{}]interface{}
	)

	for i, tc := range []struct {
		in   interface{}
		want []byte
		desc string
	}{
		{
			in:   nil,
			want: []byte("*"),
			desc: "nil",
		}, {
			in:   nullInterface,
			want: []byte("*"),
			desc: "nil interface",
		}, {
			in:   nullErr,
			want: []byte("*"),
			desc: "nil error",
		}, {
			in:   nullString,
			want: []byte("$"),
			desc: "nil string",
		}, {
			in:   nullInt,
			want: []byte("&"),
			desc: "nil int",
		}, {
			in:   nullSlice,
			want: []byte("@"),
			desc: "nil slice",
		}, {
			in:   nullMap,
			want: []byte(":"),
			desc: "nil map",
		}, {
			in:   "",
			want: []byte("$"),
			desc: "empty string",
		}, {
			in:   "kermit",
			want: []byte("$\"kermit\""),
			desc: "simple string",
		}, {
			in:   "hello world",
			want: []byte("$\"hello world\""),
			desc: "string with whitespace",
		}, {
			in:   "hi\ndu\tde!",
			want: []byte("$\"hi\\ndu\\tde!\""),
			desc: "string with control chars",
		}, {
			in:   errors.New(""),
			want: []byte("!\"unknown error\""),
			desc: "empty error",
		}, {
			in:   errors.New("some error"),
			want: []byte("!\"some error\""),
			desc: "error with text",
		}, {
			in:   errors.New("some\t\nerror"),
			want: []byte("!\"some\\t\\nerror\""),
			desc: "error with control chars",
		}, {
			in:   1,
			want: []byte("&1"),
			desc: "one dig number",
		}, {
			in:   123,
			want: []byte("&123"),
			desc: "few digs number",
		}, {
			in:   []interface{}{},
			want: []byte("@0"),
			desc: "empty slice",
		}, {
			in:   []interface{}{"kermit"},
			want: []byte("@1\n$\"kermit\""),
			desc: "one element slice",
		}, {
			in:   []interface{}{"hi", nullMap, "du\t\nde"},
			want: []byte("@3\n$\"hi\"\n:\n$\"du\\t\\nde\""),
			desc: "few element slice",
		}, {
			in:   map[interface{}]interface{}{},
			want: []byte(":0"),
			desc: "empty map",
		}, {
			in:   map[interface{}]interface{}{"hi": "dude"},
			want: []byte(":1\n$\"hi\"\n$\"dude\""),
			desc: "one key map",
		}, {
			in:   map[interface{}]interface{}{"hi": "du\nde", "some": "te\tst"},
			want: []byte(":2\n$\"hi\"\n$\"du\\nde\"\n$\"some\"\n$\"te\\tst\""),
			desc: "few element map",
		},
	} {
		got, err := Encode(tc.in)
		if err != nil {
			t.Errorf("[%d] unable to encode, error: %v", i, err)
		}

		_, ok := tc.in.(map[interface{}]interface{})
		if ok {
			// Special case for randomized map keys, have no better idea now.
			g := make([]byte, len(got))
			copy(g, got)
			w := make([]byte, len(tc.want))
			copy(w, tc.want)
			sort.Slice(g, func(i, j int) bool { return g[i] < g[j] })
			sort.Slice(w, func(i, j int) bool { return w[i] < w[j] })
			if !bytes.Equal(got, tc.want) {
				t.Errorf("[%d] %s: got %q, want %q", i, tc.desc, g, w)
			}
		}

		if !bytes.Equal(got, tc.want) {
			t.Errorf("[%d] %s: got %q, want %q", i, tc.desc, got, tc.want)
		}
	}
}

func TestEncodeUnsupported(t *testing.T) {
	_, err := Encode(int32(1))
	if err != ErrUnsupportedType {
		t.Errorf("should be unsupported, got %q, want %q", err, ErrUnsupportedType)
	}
}

func TestPrepareMessage(t *testing.T) {
	b, _ := PrepareMessage(123)
	if want := []byte("&123\r"); !bytes.Equal(b, want) {
		t.Errorf("should match format, got %q, want %q", b, want)
	}
}

func TestEncodeDecode(t *testing.T) {
	var (
		nullInterface interface{}
		nullErr       error
		nullString    string
		nullInt       int
		nullSlice     []interface{}
		nullMap       map[interface{}]interface{}
	)

	for i, tc := range []struct {
		in   interface{}
		want []byte
		desc string
	}{
		{
			in:   nil,
			want: []byte("*"),
			desc: "nil",
		}, {
			in:   nullInterface,
			want: []byte("*"),
			desc: "nil interface",
		}, {
			in:   nullErr,
			want: []byte("*"),
			desc: "nil error",
		}, {
			in:   nullString,
			want: []byte("$"),
			desc: "nil string",
		}, {
			in:   nullInt,
			want: []byte("&"),
			desc: "nil int",
		}, {
			in:   nullSlice,
			want: []byte("@"),
			desc: "nil slice",
		}, {
			in:   nullMap,
			want: []byte(":"),
			desc: "nil map",
		}, {
			in:   "",
			want: []byte("$"),
			desc: "empty string",
		}, {
			in:   "kermit",
			want: []byte("$\"kermit\""),
			desc: "simple string",
		}, {
			in:   "hello world",
			want: []byte("$\"hello world\""),
			desc: "string with whitespace",
		}, {
			in:   "hi\ndu\tde!",
			want: []byte("$\"hi\\ndu\\tde!\""),
			desc: "string with control chars",
		}, {
			in:   errors.New("some error"),
			want: []byte("!\"some error\""),
			desc: "error with text",
		}, {
			in:   errors.New("some\t\nerror"),
			want: []byte("!\"some\\t\\nerror\""),
			desc: "error with control chars",
		}, {
			in:   1,
			want: []byte("&1"),
			desc: "one dig number",
		}, {
			in:   123,
			want: []byte("&123"),
			desc: "few digs number",
		}, {
			in:   []interface{}{},
			want: []byte("@0"),
			desc: "empty slice",
		}, {
			in:   []interface{}{"kermit"},
			want: []byte("@1\n$\"kermit\""),
			desc: "one element slice",
		}, {
			in:   []interface{}{"hi", nullMap, "du\t\nde"},
			want: []byte("@3\n$\"hi\"\n:\n$\"du\\t\\nde\""),
			desc: "few element slice",
		}, {
			in:   map[interface{}]interface{}{},
			want: []byte(":0"),
			desc: "empty map",
		}, {
			in:   map[interface{}]interface{}{"hi": "dude"},
			want: []byte(":1\n$\"hi\"\n$\"dude\""),
			desc: "one key map",
		}, {
			in:   map[interface{}]interface{}{"hi": "du\nde", "some": "te\tst"},
			want: []byte(":2\n$\"hi\"\n$\"du\\nde\"\n$\"some\"\n$\"te\\tst\""),
			desc: "few element map",
		},
	} {
		encoded, err := Encode(tc.in)
		if err != nil {
			t.Errorf("[%d] unable to encode, error: %v", i, err)
		}

		r := bytes.NewReader(encoded)
		decoded, err := Decode(NewScanner(r))
		if err != nil {
			t.Errorf("[%d] unable to decode, input: %q, error: %v", i, encoded, err)
		}

		if !reflect.DeepEqual(decoded, tc.in) {
			t.Errorf("[%d] %s, encoding/decoding should be consistent: got %q, want %q", i, tc.desc, decoded, tc.in)
		}
	}
}
