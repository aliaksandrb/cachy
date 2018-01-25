package proto

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
)

type protoType byte

const (
	stringType protoType = '$'
	listType   protoType = '@'
	dictType   protoType = ':'
	errType    protoType = '!'
	nilType    protoType = '*'
)

var (
	ErrUnsupportedType = errors.New("unsupported type")
	ErrBadMsg          = errors.New("malformed message")
)

var sep = []byte{'\r', '\n'}

func Decode(in []byte) (interface{}, error) {
	if len(in) < 3 || !bytes.HasSuffix(in, sep) {
		return nil, ErrBadMsg
	}

	return decode(bytes.TrimSuffix(in, sep))
}

func decode(in []byte) (interface{}, error) {
	r := bytes.NewReader(in)
	firstByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch protoType(firstByte) {
	case stringType:
		return decodeString(r)
	case nilType:
		return nil, decodeNil(r)
	case listType:
		return decodeList(r)
	case dictType:
		return decodeDict(r)
	case errType:
		return nil, decodeErr(r)
	}

	return nil, ErrUnsupportedType
}

func decodeString(r *bytes.Reader) (string, error) {
	b := make([]byte, r.Len())
	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return string(b), nil
}

func decodeNil(r *bytes.Reader) error {
	if r.Len() > 0 {
		return ErrBadMsg
	}
	return nil
}

func decodeErr(r *bytes.Reader) error {
	s, err := decodeString(r)
	if err != nil {
		return err
	}

	if s == "" {
		return ErrBadMsg
	}

	return errors.New(s)
}

func decodeList(r *bytes.Reader) ([]interface{}, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	b := scanner.Bytes()

	size, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, ErrBadMsg
	}

	list := make([]interface{}, 0, size)
	if size == 0 {
		return list, nil
	}

	for scanner.Scan() {
		v, err := decode(scanner.Bytes())
		if err != nil {
			return nil, err
		}

		list = append(list, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(list) != cap(list) {
		return nil, ErrBadMsg
	}

	return list, nil
}

func decodeDict(r *bytes.Reader) (interface{}, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	b := scanner.Bytes()
	size, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, ErrBadMsg
	}

	dict := make(map[string]interface{}, size)
	if size == 0 {
		return dict, nil
	}

	var key string
	var assign, ok bool

	for scanner.Scan() {
		v, err := decode(scanner.Bytes())
		if err != nil {
			return nil, err
		}

		if assign {
			dict[key] = v
			assign = false
			continue
		}

		key, ok = v.(string)
		if !ok {
			return nil, ErrBadMsg
		}
		assign = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dict) != size {
		return nil, ErrBadMsg
	}

	return dict, nil
}
