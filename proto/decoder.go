package proto

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"time"

	log "github.com/aliaksandrb/cachy/logger"
)

// BytesReader interface used for blocks reading.
type BytesReader interface {
	// ReadBytes reads until the first occurrence of delim in the input.
	ReadBytes(delim byte) ([]byte, error)
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

	// TODO seek?
	b, err := buf.ReadByte()
	if err != nil {
		if err == io.EOF {
			log.Err("end of client")
			return nil, err
		}

		log.Err("unable to read first byte from message: %v", err)
		return nil, ErrBadMsg
	}

	marker := marker(b)
	mk, err := msgKindByMarker(marker)
	if err != nil {
		return nil, err
	}

	return decodeMsg(buf, mk, marker)
}

func msgKindByMarker(m marker) (mk msgKind, err error) {
	switch m {
	case CmdGet, CmdSet, CmdUpdate, CmdRemove, CmdKeys:
		return kindReq, nil
	case stringType, sliceType, mapType, errType, nilType:
		return kindRes, nil
	}

	log.Err("unknown first byte marker: %v", m)
	return mk, ErrUnsupportedCmd
}

func decodeMsg(buf *bufio.Reader, mk msgKind, m marker) (obj interface{}, err error) {
	if mk == kindReq {
		return decodeReq(buf, m)
	}

	return decodeRes(buf)
}

func decodeReq(buf *bufio.Reader, m marker) (req *Req, err error) {
	req = &Req{
		Cmd: m,
	}

	b, err := buf.ReadByte()
	if err != nil {
		log.Err("unable to read request delimiter: %v", err)
		return nil, ErrBadMsg
	}

	if b == cr && m == CmdKeys {
		return req, nil
	}

	if b != nl {
		log.Err("bad request message delimiter: %s", b)
		return nil, ErrBadDelimiter
	}

	switch m {
	case CmdGet, CmdRemove:
		return reqWithoutValue(buf, req)
	case CmdSet, CmdUpdate:
		return reqWithValue(buf, req)
	}

	log.Err("that should never happen, unsupported request command: %v", m)
	return nil, ErrUnsupportedCmd
}

func assignReqKey(br BytesReader, req *Req, d byte) error {
	b, err := br.ReadBytes(d)
	if err != nil {
		log.Err("unable to decode message key: %v", err)
		return ErrBadMsg
	}

	req.Key = string(b[:len(b)-1])

	return nil
}

func assignReqValue(buf *bufio.Reader, req *Req) error {
	val, err := decodeValue(buf)
	if err != nil {
		return err
	}

	req.Value = val

	return nil
}

func reqWithoutValue(br BytesReader, req *Req) (*Req, error) {
	err := assignReqKey(br, req, cr)
	if err != nil {
		return nil, err
	}

	return req, err
}

func reqWithValue(buf *bufio.Reader, req *Req) (*Req, error) {
	var err error

	if err = assignReqKey(buf, req, nl); err != nil {
		return nil, err
	}

	if err = assignReqValue(buf, req); err != nil {
		return nil, err
	}

	if err = assignReqTTL(buf, req); err != nil {
		return nil, err
	}

	return req, nil
}

func assignReqTTL(br BytesReader, req *Req) error {
	b, err := br.ReadBytes(cr)
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

func decodeRes(buf *bufio.Reader) (obj interface{}, err error) {
	obj, err = decodeValue(buf)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func decodeValue(buf *bufio.Reader) (interface{}, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch marker(b) {
	case stringType:
		return decodeString(buf)
	case nilType:
		return nil, decodeNil(buf)
	case sliceType:
		return decodeSlice(buf)
	case mapType:
		return decodeMap(buf)
	case errType:
		return decodeErr(buf)
	}

	log.Err("unsupported message type: %q", b)
	return nil, ErrUnsupportedType
}

func decodeString(br BytesReader) (string, error) {
	b, err := br.ReadBytes(nl)
	if err != nil {
		log.Err("unable to decode string: %q", b)
		return "", err
	}

	return string(b[:len(b)-1]), nil
}

func decodeNil(br io.ByteReader) error {
	b, err := br.ReadByte()
	if err != nil || (b != nl && b != cr) {
		log.Err("unable to decode nil: %q, err: %v", b, err)
		return ErrBadMsg
	}
	return nil
}

func decodeErr(br BytesReader) (error, error) {
	s, err := decodeString(br)
	if err != nil {
		return nil, err
	}

	// TODO check this part
	if s == "" {
		return ErrBadMsg, nil
	}

	// TODO check other errors too
	if s == ErrUnsupportedType.Error() {
		return ErrUnsupportedType, nil
	}

	return errors.New(s), nil
}

func decodeSlice(buf *bufio.Reader) (list []interface{}, err error) {
	size, err := decodeSize(buf)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return list, nil
	}

	list = make([]interface{}, 0, size)

	for i := 0; i < size; i++ {
		v, err := decodeValue(buf)
		if err != nil {
			return nil, err
		}

		list = append(list, v)
	}

	return list, nil
}

func decodeSize(br BytesReader) (size int, err error) {
	b, err := br.ReadBytes(nl)
	if err != nil {
		log.Err("unable to decode size: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	size, err = strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		log.Err("unable to convert size from string to int: %q, error: %v", b, err)
		return 0, ErrBadMsg
	}

	return size, nil
}

func decodeMap(buf *bufio.Reader) (dict map[interface{}]interface{}, err error) {
	size, err := decodeSize(buf)
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
		v, err := decodeValue(buf)
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
	ttl, err := strconv.Atoi(string(b[:len(b)-1]))
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
