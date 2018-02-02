package proto

import "io"

// Writer interface intended for writes of encoded objects to destination.
type Writer interface {
	// Write encodes obj and writes data into io.Writer w.
	// It fails only if it can't write to w.
	Write(w io.Writer, obj interface{}) error
}

// Write implements Writer interface.
func Write(w io.Writer, obj interface{}) error {
	encoded, err := Encode(obj, false)
	if err != nil {
		return WriteUnknownErr(w)
	}

	_, err = w.Write(encoded)
	return err
}

var unknownErrEncoded []byte = append([]byte(ErrUnknown.Error()), '\r', '\n')

// WriteUnknownErr writes encoded ErrUnknown to writer w.
func WriteUnknownErr(w io.Writer) error {
	_, err := w.Write(unknownErrEncoded)
	return err
}