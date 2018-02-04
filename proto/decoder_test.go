package proto

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	var (
		nullSlice []interface{}
		nullMap   map[interface{}]interface{}
	)

	for i, tc := range []struct {
		in   []byte
		want interface{}
		desc string
	}{
		{
			in:   []byte("*"),
			want: nil,
			desc: "nil",
		}, {
			in:   []byte("$"),
			want: "",
			desc: "empty string",
		}, {
			in:   []byte("&"),
			want: 0,
			desc: "empty int",
		}, {
			in:   []byte("@"),
			want: nullSlice,
			desc: "nil slice",
		}, {
			in:   []byte(":"),
			want: nullMap,
			desc: "nil map",
		}, {
			in:   []byte("$\"kermit\""),
			want: "kermit",
			desc: "simple string",
		}, {
			in:   []byte("$\"hello world\""),
			want: "hello world",
			desc: "string with whitespace",
		}, {
			in:   []byte("$\"hi\\ndu\\tde!\""),
			want: "hi\ndu\tde!",
			desc: "string with control chars",
		}, {
			in:   []byte("!\"unknown error\""),
			want: ErrUnknown,
			desc: "known error",
		}, {
			in:   []byte("!\"unsupported type\""),
			want: ErrUnsupportedType,
			desc: "known error",
		}, {
			in:   []byte("!\"unsupported command\""),
			want: ErrUnsupportedCmd,
			desc: "known error",
		}, {
			in:   []byte("!\"malformed message\""),
			want: ErrBadMsg,
			desc: "known error",
		}, {
			in:   []byte("!\"bad delimiter\""),
			want: ErrBadDelimiter,
			desc: "known error",
		}, {
			in:   []byte("!\"some error\""),
			want: errors.New("some error"),
			desc: "random error",
		}, {
			in:   []byte("!\"some\\t\\nerror\""),
			want: errors.New("some\t\nerror"),
			desc: "error with control chars",
		}, {
			in:   []byte("&1"),
			want: 1,
			desc: "one dig number",
		}, {
			in:   []byte("&123"),
			want: 123,
			desc: "few digs number",
		}, {
			in:   []byte("@0"),
			want: []interface{}{},
			desc: "empty slice",
		}, {
			in:   []byte("@1\n$\"kermit\""),
			want: []interface{}{"kermit"},
			desc: "one element slice",
		}, {
			in:   []byte("@3\n$\"hi\"\n:\n$\"du\\t\\nde\""),
			want: []interface{}{"hi", nullMap, "du\t\nde"},
			desc: "few element slice",
		}, {
			in:   []byte(":0"),
			want: map[interface{}]interface{}{},
			desc: "empty map",
		}, {
			in:   []byte(":1\n$\"hi\"\n$\"dude\""),
			want: map[interface{}]interface{}{"hi": "dude"},
			desc: "one key map",
		}, {

			in:   []byte(":2\n$\"hi\"\n$\"du\\nde\"\n$\"some\"\n$\"te\\tst\""),
			want: map[interface{}]interface{}{"hi": "du\nde", "some": "te\tst"},
			desc: "few element map",
		},
	} {
		r := bytes.NewReader(tc.in)
		got, err := Decode(NewScanner(r))
		if err != nil {
			t.Errorf("[%d] unable to decode, input: %q, error: %v", i, tc.in, err)
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("[%d] %s: got %q, want %q", i, tc.desc, got, tc.want)
		}
	}
}

func TestDecodeUnsupported(t *testing.T) {
	r := bytes.NewReader([]byte{'>'})
	_, err := Decode(NewScanner(r))
	if err != ErrUnsupportedType {
		t.Errorf("should be unsupported, got %q, want %q", err, ErrUnsupportedType)
	}

	r = bytes.NewReader([]byte(""))
	_, err = Decode(NewScanner(r))
	if err != ErrBadMsg {
		t.Errorf("should be unsupported, got %q, want %q", err, ErrBadMsg)
	}
}

//func TestDecodeMessage(t *testing.T) {
//	msg, payload, _ := DecodeMessage([]byte("+key\n$val\\nue\nk"))
//	if want := []byte("&123\r"); !bytes.Equal(b, want) {
//		t.Errorf("should match format, got %q, want %q", b, want)
//	}
//}
