package proto

import (
	"bufio"

	log "github.com/aliaksandrb/cachy/logger"
)

// Extracter used to extract raw value payload from incomming messages.
type Extracter interface {
	// Extract extracts raw value payload from a scanner s which has head+key already read.
	// It is done during DecodeMessage phase.
	// We do that to store raw value in db later.
	// It should never fail because of user input.
	Extract(s *bufio.Scanner) ([]byte, error)
}

// Extract implements Extracter.
func Extract(s *bufio.Scanner) (b []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown extraction error: %v", e)
			err = ErrUnknown
		}
	}()

	b, err = ReadBytes(s)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		log.Err("empty or malformed payload: %q", b)
		return nil, ErrBadMsg
	}

	switch b[0] {
	case STRING, INT, NIL, ERROR:
		return b, nil
	case SLICE:
		return extractSlice(b, s)
	case MAP:
		return extractMap(b, s)
	}

	log.Err("unsupported payload type: %q", b)
	return nil, ErrUnsupportedType
}

func extractSlice(head []byte, s *bufio.Scanner) ([]byte, error) {
	if len(head) == 1 {
		return head, nil
	}

	size, err := decodeSize(head[1:])
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return head, nil
	}

	for i := 0; i < size; i++ {
		v, err := Extract(s)
		if err != nil {
			return nil, err
		}

		head = append(head, NL)
		head = append(head, v...)
	}

	return head, nil
}

func extractMap(head []byte, s *bufio.Scanner) ([]byte, error) {
	if len(head) == 1 {
		return head, nil
	}

	size, err := decodeSize(head[1:])
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return head, nil
	}

	for i := 0; i < size*2; i++ {
		v, err := Extract(s)
		if err != nil {
			return nil, err
		}

		head = append(head, NL)
		head = append(head, v...)
	}

	return head, nil
}
