package proto

import (
	"bufio"
	"io"
	"strconv"
	"time"

	log "github.com/aliaksandrb/cachy/logger"
)

type head byte

const (
	CmdGet    head = '#'
	CmdSet    head = '+'
	CmdUpdate head = '^'
	CmdRemove head = '-'
	CmdKeys   head = '~'

	stringType head = '$'
	sliceType  head = '@'
	mapType    head = ':'
	errType    head = '!'
	nilType    head = '*'
)

type kind byte

const (
	kindReq kind = iota
	kindResp
)

type Request struct {
	Cmd head
	//TODO convert to bytes
	Key   string
	Value interface{}
	TTL   time.Duration
}

type Decoder interface {
	Decode(r io.Reader) (msg interface{}, err error)
}

func NewDecoder() Decoder {
	return &decoder{}
}

type decoder struct{}

func (d *decoder) Decode(r io.Reader) (val interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown decoding error: %v", e)
			err = ErrUnknown
		}
	}()

	br := bufio.NewReader(r)

	firstByte, err := br.ReadByte()
	if err != nil {
		log.Err("unable to read from a message buffer: %v", err)
		return nil, ErrBadMsg
	}

	h := head(firstByte)
	kind, err := parseHead(h)
	if err != nil {
		log.Err("unknown first byte: %v", err)
		return nil, ErrBadMsg
	}

	if kind == kindReq {
		return decodeReq(br, h)
	}

	if err = br.UnreadByte(); err != nil {
		return nil, ErrUnknown
	}

	// TODO reset buff
	// TODO get rid of kind types
	return decodeRes(br)
}

func parseHead(h head) (mk kind, err error) {
	switch h {
	case CmdGet, CmdSet, CmdUpdate, CmdRemove, CmdKeys:
		return kindReq, nil
	case stringType, sliceType, mapType, errType, nilType:
		return kindResp, nil
	}

	return mk, ErrUnsupportedType
}

const (
	nl byte = '\n'
	cr byte = '\r'
)

func decodeReq(br *bufio.Reader, cmd head) (req *Request, err error) {
	req = &Request{
		Cmd: cmd,
	}

	b, err := br.ReadByte()
	if err != nil {
		log.Err("unable to read from a message buffer: %v", err)
		return nil, ErrBadMsg
	}

	if b == cr && cmd == CmdKeys {
		return req, nil
	}

	if b != nl {
		return nil, ErrBadDelimiter
	}

	var d byte
	switch cmd {
	case CmdGet, CmdRemove:
		d = cr
	case CmdSet, CmdUpdate:
		d = nl
	default:
		return nil, ErrUnsupportedType
	}

	key, err := br.ReadBytes(d)
	if err != nil {
		log.Err("unable to decode message key: %v", err)
		return nil, ErrBadMsg
	}
	req.Key = string(key[:len(key)-1])

	if cmd == CmdGet || cmd == CmdRemove {
		return req, nil
	}

	return decodeSet(br, req)
}

func decodeSet(br *bufio.Reader, req *Request) (r *Request, err error) {
	value, err := br.ReadBytes(nl)
	if err != nil {
		log.Err("unable to decode message value: %v", err)
		return nil, ErrBadMsg
	}

	val, err := decodeValue(value[:len(value)-1])
	if err != nil {
		log.Err("unable to decode message value: %v", err)
	}

	req.Value = val

	return decodeTTL(br, req)
}

func decodeTTL(br *bufio.Reader, req *Request) (r *Request, err error) {
	ttl, err := br.ReadBytes(cr)
	if err != nil {
		log.Err("unable to decode message ttl: %v", err)
		return nil, ErrBadMsg
	}

	// TODO better way
	num, err := strconv.Atoi(string(ttl[:len(ttl)-1]))
	if err != nil {
		log.Err("bad ttl format: %v", err)
		return nil, ErrBadMsg
	}

	req.TTL = time.Duration(num)

	return req, nil
}

func decodeRes(br *bufio.Reader) (v interface{}, err error) {
	value, err := br.ReadBytes(cr)
	if err != nil {
		log.Err("unable to decode message value: %v", err)
		return nil, ErrBadMsg
	}

	return decodeValue(value[:len(value)-1])
}
