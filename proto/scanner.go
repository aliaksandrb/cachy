package proto

import (
	"bufio"
	"bytes"
	"io"
)

// NewScanner wraps incomming reader r into scaner with a per-byte split function.
func NewScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Split(onInput)
	return s
}

// NewResponseScanner reads from reader r until \r and wraps the data read into scanner.
// That scanner used for decoding, while incomming reader (which is TCPConn) is freed up
func NewResponseScanner(r io.Reader) (*bufio.Scanner, error) {
	b, err := bufio.NewReader(r).ReadBytes(CR)
	if err != nil {
		return nil, err
	}
	return NewScanner(bytes.NewReader(b)), nil
}

// BytesReader is interface for per-byte reading.
type BytesReader interface {
	// ReadBytes scans until the first occurrence of \n or \r in s.
	// It returns bytes being scanned without mentioned control sequences.
	ReadBytes(s *bufio.Scanner) (b []byte, err error)
}

// ReadBytes implements BytesReader.
func ReadBytes(s *bufio.Scanner) (b []byte, err error) {
	b = make([]byte, 0, 1)
	var read []byte

	for s.Scan() {
		read = s.Bytes()
		if len(read) == 0 || read[0] == NL || read[0] == CR {
			break
		}
		b = append(b, read[0])
	}
	if err := s.Err(); err != nil {
		return b, err
	}

	return b, nil
}

var onInput = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) != 0 && data[0] == CR {
		return 0, data, bufio.ErrFinalToken
	}

	return bufio.ScanBytes(data, atEOF)
}
