package proto

import (
	"io"
)

// Writer used to write streams of encoded data into io.Writer.
type Writer interface {
	// Write encodes obj and writes data into io.Writer w.
	// It fails only if it ca not write to w.
	Write(w io.Writer, obj interface{}) error
}

// Write implements Writer interface.
func Write(w io.Writer, obj interface{}) error {
	encoded, err := PrepareMessage(obj)
	if err != nil {
		return WriteUnknownErr(w)
	}

	_, err = w.Write(encoded)
	return err
}

// WriteRaw writes raw encoded data b into w with nil value handling.
func WriteRaw(w io.Writer, b []byte) error {
	if b == nil || len(b) == 0 {
		b = nilEnc
	}

	b = append(b, CR)

	_, err := w.Write(b)
	return err
}

var unknownErrEncoded = makeErrEncoded()

// WriteUnknownErr writes encoded ErrUnknown to writer w.
func WriteUnknownErr(w io.Writer) error {
	b := append(unknownErrEncoded, CR)
	_, err := w.Write(b)
	return err
}

func makeErrEncoded() []byte {
	return encodeErr(ErrUnknown)
}
