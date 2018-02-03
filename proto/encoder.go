package proto

import (
	"fmt"

	log "github.com/aliaksandrb/cachy/logger"
)

// Encoder interface intended for protocol messages encoding.
type Encoder interface {
	// Encode dumps obj into encoded byte slice msg.
	// It should never panic because of user input.
	Encode(obj interface{}) (msg []byte, err error)
}

func Encode(obj interface{}) (b []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown encoding error: %v", e)
			err = ErrUnknown
		}
	}()

	b, err = encode(obj)
	if err != nil {
		return b, err
	}

	b[len(b)-1] = cr
	return append(b, nl), nil
}

func encode(obj interface{}) (b []byte, err error) {
	switch t := obj.(type) {
	case nil:
		return encodeNil(), nil
	case error:
		return encodeErr(t), nil
	case string:
		return encodeString(t), nil
	case []interface{}:
		return encodeSlice(t), nil
	case []string:
		return encodeStringSlice(t), nil
	case map[interface{}]interface{}:
		return encodeMap(t), nil
	}

	log.Err("unknown obj to encode: %v", obj)
	return nil, ErrUnsupportedType
}

var nilEnc = []byte{byte(nilType), nl}

func encodeNil() []byte {
	return nilEnc
}

func encodeErr(in error) []byte {
	res := encodeString(in.Error())
	res[0] = byte(errType)
	return res
}

func encodeString(in string) []byte {
	msg := []byte(in)
	l := len(msg)

	res := make([]byte, l+2, l+3)
	res[0] = byte(stringType)
	copy(res[1:], msg)
	res[l+1] = nl

	return res
}

func encodeSlice(in []interface{}) []byte {
	res := collectionHeader(len(in))
	res[0] = byte(sliceType)

	for _, v := range in {
		vals, err := encode(v)
		if err != nil {
			return encodeErr(err)
		}

		res = append(res, vals...)
	}

	return res
}

func collectionHeader(size int) []byte {
	length := []byte(fmt.Sprintf("%d", size))
	l := len(length)

	res := make([]byte, l+2, 3+l+size*3)
	copy(res[1:], length)
	res[l+1] = nl

	return res
}

func encodeMap(in map[interface{}]interface{}) []byte {
	res := collectionHeader(len(in))
	res[0] = byte(mapType)

	for k, v := range in {
		key, err := encode(k)
		if err != nil {
			return encodeErr(err)
		}
		res = append(res, key...)

		value, err := encode(v)
		if err != nil {
			return encodeErr(err)
		}
		res = append(res, value...)
	}

	return res
}

func encodeStringSlice(in []string) []byte {
	res := collectionHeader(len(in))
	res[0] = byte(sliceType)

	for _, v := range in {
		vals := encodeString(v)
		res = append(res, vals...)
	}

	return res
}
