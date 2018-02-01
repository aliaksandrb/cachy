package proto

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/aliaksandrb/cachy/logger"
)

var (
	sep    = []byte{'\r', '\n'}
	nilEnc = []byte{byte(nilType), '\r', '\n'}
	//emptyStrEnc = []byte{'$', '\r', '\n'}
)

// TODO make skipEnding optional
// TODO accept buff
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
	case map[string]interface{}:
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

func encodeMap(in map[string]interface{}, skipEnding bool) []byte {
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

//func Decode(in []byte) (interface{}, error) {
//	if len(in) < 3 || !bytes.HasSuffix(in, sep) {
//		return nil, ErrBadMsg
//	}
//
//	return decode(bytes.TrimSuffix(in, sep))
//}

func decodeValue(in []byte) (interface{}, error) {
	//in = bytes.Replace(in, []byte("\\n"), []byte{'\n'}, -1)

	log.Info("decodeValue %q", in)
	r := bytes.NewReader(in)
	firstByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	log.Info("firstByte %q", firstByte)

	switch marker(firstByte) {
	case stringType:
		return decodeString(r)
	case nilType:
		return nil, decodeNil(r)
	case sliceType:
		return decodeSlice(r)
	case mapType:
		return decodeMap(r)
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

	if s == ErrUnsupportedType.Error() {
		return ErrUnsupportedType
	}

	return errors.New(s)
}

func decodeSlice(r *bytes.Reader) ([]interface{}, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	b := scanner.Bytes()

	log.Info("decodeSlice %q", b)

	size, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, ErrBadMsg
	}

	list := make([]interface{}, 0, size)
	if size == 0 {
		return list, nil
	}

	for scanner.Scan() {
		v, err := decodeValue(scanner.Bytes())
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

func decodeMap(r *bytes.Reader) (interface{}, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	if !scanner.Scan() {
		return nil, ErrBadMsg
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

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
		v, err := decodeValue(scanner.Bytes())
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
