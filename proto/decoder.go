package proto

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"time"

	log "github.com/aliaksandrb/cachy/logger"
)

// BytesReader interface used for per-byte reading.
type BytesReader interface {
	// ReadBytes reads until the first occurrence of \n or \r in the underlaying reader in s.
	ReadBytes(s *bufio.Scanner) (b []byte, err error)
}

// ReadBytes scans s until first occurance of \n or \r and returns
// the byte slice of data being scanned.
func ReadBytes(s *bufio.Scanner) (b []byte, err error) {
	b = make([]byte, 0, 1)
	var read []byte

	for s.Scan() {
		read = s.Bytes()
		if len(read) == 0 || read[0] == nl || read[0] == cr {
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
	if len(data) != 0 && data[0] == cr {
		return 0, data, bufio.ErrFinalToken
	}

	return bufio.ScanBytes(data, atEOF)
}

// Decoder interface intended for protocol messages parsing.
type Decoder interface {
	// Decode reads from buffer buf and returns decoded obj and error err if any.
	// It should never panic because of user input.
	Decode(buf *bufio.Reader) (obj interface{}, err error)
}

// Decode implements Decoder interface.
func Decode(buf *bufio.Reader) (obj interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown decoding error: %v", e)
			obj, err = nil, ErrUnknown
		}
	}()

	b, err := buf.Peek(1)
	if err != nil {
		if err == io.EOF {
			log.Err("end of client")
			return nil, err
		}

		log.Err("unable to read first byte from message: %v", err)
		return nil, ErrBadMsg
	}

	marker := marker(b[0])
	mk, err := msgKindByMarker(marker)
	if err != nil {
		return nil, err
	}

	return decode(buf, mk, marker)
}

func msgKindByMarker(m marker) (mk msgKind, err error) {
	switch m {
	case CmdGet, CmdSet, CmdUpdate, CmdRemove, CmdKeys:
		return kindReq, nil
	case stringType, sliceType, mapType, errType, nilType:
		return kindRes, nil
	}

	log.Err("unknown first byte marker: %q", m)
	return mk, ErrUnsupportedCmd
}

func decode(buf *bufio.Reader, mk msgKind, m marker) (obj interface{}, err error) {
	s := bufio.NewScanner(buf)
	s.Split(onInput)

	if mk == kindReq {
		return decodeReq(s, m)
	}

	return decodeRes(s)
}

func decodeReq(s *bufio.Scanner, m marker) (req *Req, err error) {
	req = &Req{
		Cmd: m,
	}

	log.Info("decodeReq: %q", m)

	b, err := ReadBytes(s)
	if err != nil || len(b) == 0 {
		log.Err("unable to read request delimiter: %q, err: %v", b, err)
		return nil, ErrBadMsg
	}

	switch m {
	case CmdKeys:
		return req, nil
	case CmdGet, CmdRemove:
		return reqWithoutValue(s, req)
	case CmdSet, CmdUpdate:
		return reqWithValue(s, req)
	}

	log.Err("that should never happen, unsupported request command: %q", m)
	return nil, ErrUnsupportedCmd
}

func assignReqKey(s *bufio.Scanner, req *Req, d byte) error {
	// TODO decode key?
	b, err := ReadBytes(s)
	if err != nil {
		log.Err("unable to decode message key: %v", err)
		return ErrBadMsg
	}

	req.Key = string(b)

	return nil
}

func assignReqValue(s *bufio.Scanner, req *Req) error {
	val, err := decodeValue(s)
	if err != nil {
		return err
	}

	req.Value = val

	return nil
}

func reqWithoutValue(s *bufio.Scanner, req *Req) (*Req, error) {
	err := assignReqKey(s, req, cr)
	if err != nil {
		return nil, err
	}

	return req, err
}

func reqWithValue(s *bufio.Scanner, req *Req) (*Req, error) {
	var err error
	log.Info("reqWithValue1: %+v", req)

	if err = assignReqKey(s, req, nl); err != nil {
		return nil, err
	}

	log.Info("reqWithValue2: %+v", req)

	if err = assignReqValue(s, req); err != nil {
		return nil, err
	}

	log.Info("reqWithValue3: %+v", req)
	if err = assignReqTTL(s, req); err != nil {
		return nil, err
	}
	log.Info("reqWithValue4: %+v", req)

	return req, nil
}

func assignReqTTL(s *bufio.Scanner, req *Req) error {
	b, err := ReadBytes(s)
	if err != nil {
		log.Err("unable to decode message ttl: %v", err)
		return ErrBadMsg
	}

	ttl, err := bytesToDuration(b)
	if err != nil {
		return err
	}

	req.TTL = ttl

	return nil
}

func decodeRes(s *bufio.Scanner) (obj interface{}, err error) {
	log.Info("decodeRes")

	obj, err = decodeValue(s)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func decodeValue(s *bufio.Scanner) (interface{}, error) {
	log.Info("decodeValue")
	b, err := ReadBytes(s)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		log.Err("empty value marker: %q", b)
		return nil, ErrBadMsg
	}

	log.Info("switch %q - %q", b, b[0])
	switch marker(b[0]) {
	case stringType:
		return decodeString(s)
	case nilType:
		return nil, decodeNil(s)
	case sliceType:
		return decodeSlice(s)
	case mapType:
		return decodeMap(s)
	case errType:
		return decodeErr(s)
	}

	log.Err("unsupported message type: %q", b)
	return nil, ErrUnsupportedType
}

func decodeString(s *bufio.Scanner) (string, error) {
	log.Info("decodeString")
	b, err := ReadBytes(s)
	if err != nil {
		return "", err
	}

	log.Info("decodeString: %q", b)
	return string(b), nil
}

func decodeNil(s *bufio.Scanner) error {
	b, err := ReadBytes(s)
	if err != nil || (len(b) != 1 && b[0] != nl && b[0] != cr) {
		log.Err("unable to decode nil: %q, err: %v", b, err)
		return ErrBadMsg
	}
	return nil
}

func decodeErr(s *bufio.Scanner) (error, error) {
	str, err := decodeString(s)
	if err != nil {
		return nil, err
	}

	if str == "" {
		return ErrBadMsg, nil
	}

	// TODO check other errors too
	if str == ErrUnsupportedType.Error() {
		return ErrUnsupportedType, nil
	}

	return errors.New(str), nil
}

func decodeSlice(s *bufio.Scanner) (list []interface{}, err error) {
	size, err := decodeSize(s)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return list, nil
	}

	list = make([]interface{}, 0, size)

	for i := 0; i < size; i++ {
		v, err := decodeValue(s)
		if err != nil {
			return nil, err
		}

		list = append(list, v)
	}

	return list, nil
}

func decodeSize(s *bufio.Scanner) (size int, err error) {
	b, err := ReadBytes(s)
	if err != nil {
		log.Err("unable to decode size: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	size, err = strconv.Atoi(string(b))
	if err != nil {
		log.Err("unable to convert size from string to int: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	return size, nil
}

func decodeMap(s *bufio.Scanner) (dict map[interface{}]interface{}, err error) {
	size, err := decodeSize(s)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return dict, nil
	}

	dict = make(map[interface{}]interface{}, size)

	var key interface{}
	var assign bool

	for i := 0; i < size*2; i++ {
		v, err := decodeValue(s)
		if err != nil {
			return nil, err
		}

		if assign {
			dict[key] = v
			assign = false
			continue
		}

		key = v
		assign = true
	}

	return dict, nil
}

func bytesToDuration(b []byte) (t time.Duration, err error) {
	// TODO better way
	ttl, err := strconv.Atoi(string(b))
	if err != nil {
		log.Err("bad ttl format: %v", err)
		return
	}

	if ttl < 0 {
		log.Err("negative ttl doesn't make sense: %v", ttl)
		err = ErrBadMsg
		return
	}

	return time.Duration(ttl), nil
}
