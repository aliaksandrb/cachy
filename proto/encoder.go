package proto

import (
	"fmt"
	"strconv"

	log "github.com/aliaksandrb/cachy/logger"
)

// Encoder interface used to encode runtime objects into protocol format.
type Encoder interface {
	// Encode encodes obj into protocol byte slice b.
	// It should never panic because of user input.
	Encode(obj interface{}) (b []byte, err error)
	// PrepareMessage prepares ready to be send over network message b for provided obj.
	PrepareMessage(obj interface{}) (b []byte, err error)
}

// PrepareMessage implements Encoder interface.
func PrepareMessage(obj interface{}) (b []byte, err error) {
	b, err = Encode(obj)
	if err != nil {
		return
	}

	return append(b, CR), nil
}

// Encode implements Encoder interface.
func Encode(obj interface{}) (b []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Err("unknown encoding error: %v", e)
			err = ErrUnknown
		}
	}()

	switch t := obj.(type) {
	case nil:
		return encodeNil(), nil
	case string:
		return encodeString(t), nil
	case int:
		return encodeInt(t), nil
	case []interface{}:
		return encodeSlice(t)
	case []string:
		return encodeStringSlice(t)
	case map[interface{}]interface{}:
		return encodeMap(t)
	case error:
		return encodeErr(t), nil
	}

	log.Err("unknown obj type to encode: %T - %q", obj, obj)
	return nil, ErrUnsupportedType
}

var nilEnc = []byte{NIL}

func encodeNil() []byte {
	return nilEnc
}

var strEnc = []byte{STRING}

func encodeString(in string) []byte {
	if len(in) == 0 {
		return strEnc
	}

	return append(strEnc, strconv.QuoteToASCII(in)...)
}

var errEnc = []byte{ERROR}

func encodeErr(in error) []byte {
	if in == nil {
		return errEnc
	}

	msg := in.Error()
	if msg == "" {
		return append(errEnc, strconv.QuoteToASCII(ErrUnknown.Error())...)
	}

	return append(errEnc, strconv.QuoteToASCII(msg)...)
}

var intEnc = []byte{INT}

func encodeInt(in int) []byte {
	if in == 0 {
		return intEnc
	}

	return append(intEnc, IntToBytes(in)...)
}

var sliceEnc = []byte{SLICE}

func encodeSlice(in []interface{}) ([]byte, error) {
	if in == nil {
		return sliceEnc, nil
	}

	b := append(sliceEnc, IntToBytes(len(in))...)

	if len(in) == 0 {
		return b, nil
	}

	for _, v := range in {
		b = append(b, NL)

		encoded, err := Encode(v)
		if err != nil {
			return nil, err
		}

		b = append(b, encoded...)
	}

	return b, nil
}

func encodeStringSlice(in []string) ([]byte, error) {
	slice := make([]interface{}, len(in))
	for i, v := range in {
		slice[i] = v
	}

	return encodeSlice(slice)
}

func IntToBytes(i int) []byte {
	//TODO strconv
	return []byte(fmt.Sprintf("%d", i))
}

var mapEnc = []byte{MAP}

func encodeMap(in map[interface{}]interface{}) ([]byte, error) {
	if in == nil {
		return mapEnc, nil
	}

	b := append(mapEnc, IntToBytes(len(in))...)

	if len(in) == 0 {
		return b, nil
	}

	for k, v := range in {
		b = append(b, NL)
		key, err := Encode(k)
		if err != nil {
			return nil, err
		}
		b = append(b, key...)

		b = append(b, NL)
		val, err := Encode(v)
		if err != nil {
			return nil, err
		}
		b = append(b, val...)
	}

	return b, nil
}
