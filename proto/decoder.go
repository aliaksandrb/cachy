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

// Decoder interface used to decode protocol format into runtime objects.
type Decoder interface {
	// Decode reads from buffer buf and returns decoded obj and error err if any.
	// It should never panic because of user input.
	Decode(buf *bufio.Reader) (obj interface{}, err error)
}

// Decode implements Decoder interface.
func DecodeMessage(buf *bufio.Reader) (obj interface{}, err error) {
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

	marker := b[0]
	mk, err := msgKindByMarker(marker)
	if err != nil {
		return nil, err
	}

	return decode(buf, mk, marker)
}

func msgKindByMarker(m byte) (mk byte, err error) {
	switch m {
	case CmdGet, CmdSet, CmdUpdate, CmdRemove, CmdKeys:
		return KindReq, nil
	case STRING, SLICE, MAP, ERROR, NIL:
		return KindRes, nil
	}

	log.Err("unknown first byte marker: %q", m)
	return mk, ErrUnsupportedCmd
}

func NewScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Split(onInput)
	return s
}

func decode(buf *bufio.Reader, mk byte, m byte) (obj interface{}, err error) {
	s := NewScanner(buf)
	if mk == KindReq {
		return decodeReq(s, m)
	}

	return decodeRes(s)
}

func decodeReq(s *bufio.Scanner, m byte) (req *Req, err error) {
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
	val, err := Decode(s)
	if err != nil {
		return err
	}

	req.Value = val

	return nil
}

func reqWithoutValue(s *bufio.Scanner, req *Req) (*Req, error) {
	err := assignReqKey(s, req, CR)
	if err != nil {
		return nil, err
	}

	return req, err
}

func reqWithValue(s *bufio.Scanner, req *Req) (*Req, error) {
	var err error
	log.Info("reqWithValue1: %+v", req)

	if err = assignReqKey(s, req, NL); err != nil {
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

	obj, err = Decode(s)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func Decode(s *bufio.Scanner) (interface{}, error) {
	b, err := ReadBytes(s)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		log.Err("empty or malformed payload: %q", b)
		return nil, ErrBadMsg
	}

	switch b[0] {
	case STRING:
		return decodeString(b)
	case INT:
		return decodeInt(b)
	case NIL:
		return nil, decodeNil(s)
	case SLICE:
		return decodeSlice(b, s)
	case MAP:
		return decodeMap(b, s)
	case ERROR:
		return decodeErr(b)
	}

	log.Err("unsupported payload type: %q", b)
	return nil, ErrUnsupportedType
}

func decodeString(b []byte) (s string, err error) {
	if len(b) == 1 {
		return
	}

	return strconv.Unquote(string(b[1:]))
}

func decodeInt(b []byte) (i int, err error) {
	if len(b) == 1 {
		return
	}

	return decodeSize(b[1:])
}

func decodeNil(s *bufio.Scanner) error {
	log.Info("decode nil")
	b, err := ReadBytes(s)
	log.Info("decode nil %q - %d - %v", b, len(b), err)
	if err != nil { // || len(b) != 1 || (b[0] != nl && b[0] != cr) {
		log.Err("unable to decode nil: %q, err: %v", b, err)
		return ErrBadMsg
	}
	log.Info("decodeNil success")
	return nil
}

func decodeErr(b []byte) (error, error) {
	str, err := decodeString(b)
	if err != nil {
		return nil, err
	}

	if str == "" {
		return ErrBadMsg, nil
	}

	switch str {
	case ErrUnsupportedType.Error():
		return ErrUnsupportedType, nil
	case ErrUnsupportedCmd.Error():
		return ErrUnsupportedCmd, nil
	case ErrBadMsg.Error():
		return ErrBadMsg, nil
	case ErrBadDelimiter.Error():
		return ErrBadDelimiter, nil
	case ErrUnknown.Error():
		return ErrUnknown, nil
	}

	return errors.New(str), nil
}

func decodeSlice(head []byte, s *bufio.Scanner) (slice []interface{}, err error) {
	if len(head) == 1 {
		return
	}

	size, err := decodeSize(head[1:])
	if err != nil {
		return
	}

	if size == 0 {
		return []interface{}{}, nil
	}

	slice = make([]interface{}, 0, size)

	for i := 0; i < size; i++ {
		v, err := Decode(s)
		if err != nil {
			return nil, err
		}

		slice = append(slice, v)
	}

	return
}

func decodeSize(b []byte) (size int, err error) {
	// TODO better way
	size, err = strconv.Atoi(string(b))
	if err != nil {
		log.Err("unable to convert size from bytes to int: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	return
}

func decodeMap(head []byte, s *bufio.Scanner) (dict map[interface{}]interface{}, err error) {
	if len(head) == 1 {
		return
	}

	size, err := decodeSize(head[1:])
	if err != nil {
		return
	}

	if size == 0 {
		return make(map[interface{}]interface{}, 0), nil
	}

	dict = make(map[interface{}]interface{}, size)

	var key interface{}
	var assign bool

	for i := 0; i < size*2; i++ {
		v, err := Decode(s)
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
	ttl, err := decodeSize(b)
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
