package proto

import "io"

// Encoder interface intended for protocol messages encoding.
type Encoder interface {
	// Encode dumps sends data from io.Reader r to io.Writer w.
	// It should never panic because of user input.
	Encode(r io.Reader, w io.Writer) error
}

func EncodeErr(w io.Writer, e error) error {
	errEncoded, err := Encode(e, false)
	if err != nil {
		return EncodeUnknownErr(w)
	}

	_, err = w.Write(errEncoded)
	return err
}

var unknownErrEncoded []byte = append([]byte(ErrUnknown.Error()), '\r', '\n')

func EncodeUnknownErr(w io.Writer) error {
	_, err := w.Write(unknownErrEncoded)
	return err
}
