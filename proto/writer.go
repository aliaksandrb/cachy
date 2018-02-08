package proto

import (
	"io"
)

func NewWriter() *writer {
	return new(writer)
}

type writer struct{}

// Write implements Writer interface.
func (wr *writer) Write(w io.Writer, obj interface{}) error {
	encoded, err := PrepareMessage(obj)
	if err != nil {
		return wr.WriteUnknownErr(w)
	}

	_, err = w.Write(encoded)
	return err
}

// WriteRaw writes raw encoded data b into w with nil value handling.
func (wr *writer) WriteRaw(w io.Writer, b []byte) error {
	if b == nil || len(b) == 0 {
		b = nilEnc
	}

	b = append(b, CR)

	_, err := w.Write(b)
	return err
}

var unknownErrEncoded = makeErrEncoded()

// WriteUnknownErr writes encoded ErrUnknown to writer w.
func (wr *writer) WriteUnknownErr(w io.Writer) error {
	b := append(unknownErrEncoded, CR)
	_, err := w.Write(b)
	return err
}

func makeErrEncoded() []byte {
	return encodeErr(ErrUnknown)
}
