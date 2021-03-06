package proto

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"time"

	log "github.com/aliaksandrb/cachy/logger"
)

func NewDecoder() *decoder {
	return new(decoder)
}

// Decoder used to decode protocol format encoding into runtime objects.
type Decoder interface {
	// Decode reads from buffer under scanner s and returns decoded runtime obj and error err if any.
	// It should never panic because of user input.
	Decode(s *bufio.Scanner) (obj interface{}, err error)
}

type decoder struct{}

// Decode implements Decoder interface.
func (d *decoder) Decode(s *bufio.Scanner) (obj interface{}, err error) {
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
		return d.decodeSlice(b, s)
	case MAP:
		return d.decodeMap(b, s)
	case ERROR:
		return decodeErr(b)
	}

	log.Err("unsupported payload type: %q", b)
	return nil, ErrUnsupportedType
}

// DecodeMessage implements MessageDecoder interface.
func (d *decoder) DecodeMessage(buf *bufio.Reader) (m interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown decoding error: %v", e)
			err = ErrUnknown
		}
	}()

	b, err := buf.Peek(1)
	if err == io.EOF {
		log.Err("end of client")
		return nil, err
	}

	if err != nil {
		log.Err("unable to read first byte of message: %v", err)
		return nil, ErrBadMsg
	}

	marker := b[0]
	mk, err := msgKindByMarker(marker)
	if err != nil {
		return nil, err
	}

	return d.decode(buf, mk, marker)
}

func msgKindByMarker(m byte) (mk byte, err error) {
	switch m {
	case CmdGet, CmdSet, CmdUpdate, CmdRemove, CmdKeys:
		return KindReq, nil
	case STRING, INT, SLICE, MAP, ERROR, NIL:
		return KindRes, nil
	}

	log.Err("unknown first byte marker: %q", m)
	return mk, ErrUnsupportedCmd
}

func (d *decoder) decode(buf *bufio.Reader, mk byte, m byte) (obj interface{}, err error) {
	s := NewScanner(buf)

	if mk == KindReq {
		return decodeReq(s, m)
	}

	return d.Decode(s)
}

func decodeReq(s *bufio.Scanner, m byte) (req *Req, err error) {
	req = &Req{
		Cmd: m,
	}

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

func reqWithoutValue(s *bufio.Scanner, req *Req) (*Req, error) {
	err := assignReqKey(s, req)
	if err != nil {
		return nil, err
	}

	return req, err
}

func assignReqKey(s *bufio.Scanner, req *Req) error {
	b, err := ReadBytes(s)
	if err != nil {
		log.Err("unable to decode message key: %v", err)
		return ErrBadMsg
	}

	if len(b) == 0 {
		log.Err("unable to decode messge key: %v", b)
		return ErrBadMsg
	}

	req.Key = string(b)

	return nil
}

func reqWithValue(s *bufio.Scanner, req *Req) (*Req, error) {
	var err error

	if err = assignReqKey(s, req); err != nil {
		return nil, err
	}

	if err = assignReqValue(s, req); err != nil {
		return nil, err
	}

	if err = assignReqTTL(s, req); err != nil {
		return nil, err
	}

	return req, nil
}

func assignReqValue(s *bufio.Scanner, req *Req) error {
	b, err := Extract(s)
	if err != nil {
		return err
	}

	req.Value = b

	return nil
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
	b, err := ReadBytes(s)

	if err != nil {
		log.Err("unable to decode nil: %q, err: %v", b, err)
		return ErrBadMsg
	}

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

func (d *decoder) decodeSlice(head []byte, s *bufio.Scanner) (slice []interface{}, err error) {
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
		v, err := d.Decode(s)
		if err != nil {
			return nil, err
		}

		slice = append(slice, v)
	}

	return
}

func decodeSize(b []byte) (int, error) {
	size64, err := strconv.ParseInt(string(b), 10, 0)
	if err != nil {
		log.Err("unable to convert size from bytes to int: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	// Safe to do because ParseInt uses bitSize = 0
	return int(size64), nil
}

func (d *decoder) decodeMap(head []byte, s *bufio.Scanner) (dict map[interface{}]interface{}, err error) {
	if len(head) == 1 {
		return
	}

	size, err := decodeSize(head[1:])
	if err != nil {
		return
	}

	dict = make(map[interface{}]interface{}, size)

	if size == 0 {
		return dict, nil
	}

	var key interface{}
	var assign bool

	for i := 0; i < size*2; i++ {
		v, err := d.Decode(s)
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
	ttl, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		log.Err("unable to convert size from bytes to int: %q, error: %v", b, err)
		return
	}

	if ttl < 0 {
		log.Err("negative ttl doesn't make sense: %v", ttl)
		err = ErrBadMsg
		return
	}

	return time.Duration(ttl), nil
}
