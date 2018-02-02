package proto

import (
	"fmt"
)

// Encoder interface intended for protocol messages encoding.
type Encoder interface {
	// Encode dumps obj into encoded byte slice msg.
	// It should never panic because of user input.
	Encode(obj interface{}) (msg []byte, err error)
}

// TODO make skipEnding optional
func Encode(v interface{}, skipEnding bool) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return encodeNil(skipEnding), nil
	case error:
		return encodeErr(t, skipEnding), nil
	case string:
		return encodeString(t, skipEnding), nil
	case []interface{}:
		return encodeSlice(t, skipEnding), nil
	case map[interface{}]interface{}:
		return encodeMap(t, skipEnding), nil
	default:
		fmt.Printf("===> %v --- %T\n", t, t)
	}

	return nil, ErrUnsupportedType
}

func encodeNil(skipEnding bool) []byte {
	if skipEnding {
		return nilEnc[:1]
	}

	return nilEnc
}

func encodeErr(in error, skipEnding bool) []byte {
	msg := []byte(in.Error())
	res := make([]byte, len(msg)+1, len(msg)+3)
	res[0] = byte(errType)

	for i, b := range msg {
		res[i+1] = b
	}
	if !skipEnding {
		res = append(res, sep...)
	}

	return res
}

func encodeString(in string, skipEnding bool) []byte {
	msg := []byte(in)
	res := make([]byte, len(msg)+1, len(msg)+3)
	res[0] = byte(stringType)

	for i, b := range msg {
		res[i+1] = b
	}
	if !skipEnding {
		res = append(res, sep...)
	}

	return res
}

func encodeSlice(in []interface{}, skipEnding bool) []byte {
	size := len(in)
	length := []byte(fmt.Sprintf("%d", size))
	res := make([]byte, len(length)+2, 3+len(length)+size*3)

	res[0] = byte(sliceType)
	for i, b := range length {
		res[i+1] = b
	}
	res = append(res, '\n')

	for i, v := range in {
		vals, err := Encode(v, true)
		if err != nil {
			return encodeErr(err, false)
		}

		res = append(res, vals...)
		if i != size-1 {
			res = append(res, '\n')
		}
	}

	if !skipEnding {
		res = append(res, sep...)
	}

	return res
}

func encodeMap(in map[interface{}]interface{}, skipEnding bool) []byte {
	size := len(in)
	length := []byte(fmt.Sprintf("%d", size))
	res := make([]byte, len(length)+2, 3+len(length)+size*3)

	res[0] = byte(mapType)
	for i, b := range length {
		res[i+1] = b
	}
	res = append(res, '\n')

	i := 0
	for k, v := range in {
		i++
		key, err := Encode(k, true)
		if err != nil {
			return encodeErr(err, false)
		}

		res = append(res, key...)
		res = append(res, '\n')

		value, err := Encode(v, true)
		if err != nil {
			return encodeErr(err, false)
		}

		res = append(res, value...)
		if i != size-1 {
			res = append(res, '\n')
		}
	}

	if !skipEnding {
		res = append(res, sep...)
	}

	return res
}
