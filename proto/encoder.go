package proto

import "io"

// Encoder interface intended for protocol messages encoding.
type Encoder interface {
	// Encode dumps sends data from io.Reader r to io.Writer w.
	// It should never panic because of user input.
	Encode(r io.Reader, w io.Writer) error
}
