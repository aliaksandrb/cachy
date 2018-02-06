package proto

import (
	"bytes"
	"testing"
)

func TestExtract(t *testing.T) {
	for i, tc := range []struct {
		in   []byte
		want []byte
		desc string
	}{
		{
			in:   []byte("*\n0\r"),
			want: []byte("*"),
			desc: "nil",
		}, {
			in:   []byte("$\n0\r"),
			want: []byte("$"),
			desc: "empty string",
		}, {
			in:   []byte("&\n0\r"),
			want: []byte("&"),
			desc: "empty int",
		}, {
			in:   []byte("@\n0\r"),
			want: []byte("@"),
			desc: "nil slice",
		}, {
			in:   []byte(":\n0\r"),
			want: []byte(":"),
			desc: "nil map",
		}, {
			in:   []byte("!\n1\r"),
			want: []byte("!"),
			desc: "empty error",
		}, {
			in:   []byte("$\"kermit\"\n0\r"),
			want: []byte("$\"kermit\""),
			desc: "simple string",
		}, {
			in:   []byte("$\"hello world\"\n100\r"),
			want: []byte("$\"hello world\""),
			desc: "string with ttl",
		}, {
			in:   []byte("$\"hi\\ndu\\tde!\"\n0\r"),
			want: []byte("$\"hi\\ndu\\tde!\""),
			desc: "string with control chars",
		}, {
			in:   []byte("!\"some error\"\n1\r"),
			want: []byte("!\"some error\""),
			desc: "random error",
		}, {
			in:   []byte("!\"some\\t\\nerror\"\n0\r"),
			want: []byte("!\"some\\t\\nerror\""),
			desc: "error with control chars",
		}, {
			in:   []byte("&1\n0\r"),
			want: []byte("&1"),
			desc: "one dig number",
		}, {
			in:   []byte("&123\n100\r"),
			want: []byte("&123"),
			desc: "few digs number with ttl",
		}, {
			in:   []byte("@0\n0\r"),
			want: []byte("@0"),
			desc: "empty slice",
		}, {
			in:   []byte("@1\n$\"kermit\"\n100\r"),
			want: []byte("@1\n$\"kermit\""),
			desc: "one element slice",
		}, {
			in:   []byte("@3\n$\"hi\"\n:\n$\"du\\t\\nde\"\n666\r"),
			want: []byte("@3\n$\"hi\"\n:\n$\"du\\t\\nde\""),
			desc: "few element slice",
		}, {
			in:   []byte(":0\n0\r"),
			want: []byte(":0"),
			desc: "empty map",
		}, {
			in:   []byte(":1\n$\"hi\"\n$\"dude\"\n1\r"),
			want: []byte(":1\n$\"hi\"\n$\"dude\""),
			desc: "one key map",
		}, {

			in:   []byte(":2\n$\"hi\"\n$\"du\\nde\"\n$\"some\"\n$\"te\\tst\"\n42\r"),
			want: []byte(":2\n$\"hi\"\n$\"du\\nde\"\n$\"some\"\n$\"te\\tst\""),
			desc: "few element map",
		},
	} {
		r := bytes.NewReader(tc.in)
		got, err := Extract(NewScanner(r))
		if err != nil {
			t.Errorf("[%d] unable to extract, input: %q, error: %v", i, tc.in, err)
			continue
		}

		if !bytes.Equal(got, tc.want) {
			t.Errorf("[%d] %s: got %q, want %q", i, tc.desc, got, tc.want)
		}
	}
}

func TestExtractUnsupported(t *testing.T) {
	r := bytes.NewReader([]byte(">hello\n1\r"))

	_, err := Extract(NewScanner(r))
	if err != ErrUnsupportedType {
		t.Errorf("should be unsupported, got %q, want %q", err, ErrUnsupportedType)
	}

	r = bytes.NewReader([]byte("\r"))
	_, err = Extract(NewScanner(r))
	if err != ErrBadMsg {
		t.Errorf("should be unsupported, got %q, want %q", err, ErrBadMsg)
	}
}
