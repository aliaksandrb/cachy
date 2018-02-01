package proto

import (
	"bufio"
	"io"

	log "github.com/aliaksandrb/cachy/logger"
)

// BytesReader interface used for blocks reading.
type BytesReader interface {
	// ReadBytes reads until the first occurrence of delim in the input.
	ReadBytes(delim byte) ([]byte, error)
}

// Decoder interface intended for protocol messages parsing.
type Decoder interface {
	// Decode reads from buffer buf and returns proto.Request req and error err if any.
	// It should never panic because of user input.
	Decode(buf *bufio.Reader) (req Request, err error)
}

// Decode implements Decoder interface.
func Decode(buf *bufio.Reader) (req Request, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown decoding error: %v", e)
			req, err = nil, ErrUnknown
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

func decodeMsg(buf *bufio.Reader, mk msgKind, m marker) (req Request, err error) {
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

func assignReqValue(br BytesReader, req *Req) error {
	b, err := br.ReadBytes(nl)
	if err != nil {
		log.Err("unable to decode message value: %v", err)
		return ErrBadMsg
	}

	// TODO not so good to load the full value into memory.
	val, err := decodeValue(b[:len(b)-1])
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

func reqWithValue(br BytesReader, req *Req) (*Req, error) {
	var err error

	if err = assignReqKey(br, req, nl); err != nil {
		return nil, err
	}

	if err = assignReqValue(br, req); err != nil {
		return nil, err
	}

	if err = assignReqTTL(br, req); err != nil {
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

func decodeRes(br BytesReader) (res *Res, err error) {
	b, err := br.ReadBytes(cr)
	if err != nil {
		log.Err("unable to read response message value: %v", err)
		return nil, ErrBadMsg
	}

	// TODO not so good to load the full value into memory.
	val, err := decodeValue(b[:len(b)-1])
	if err != nil {
		return nil, err
	}

	return &Res{Value: val}, nil
}
