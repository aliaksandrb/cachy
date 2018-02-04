package proto

import (
	"time"
)

func NewMessage(cmd byte, key string, value interface{}, ttl time.Duration) (b []byte, err error) {
	b = []byte{cmd, NL}

	if cmd == CmdKeys {
		return append(b, CR), nil
	}

	b = append(b, []byte(key)...)

	if cmd == CmdGet || cmd == CmdRemove {
		return append(b, CR), nil
	}

	b = append(b, NL)
	valueEnc, err := Encode(value)
	if err != nil {
		return nil, err
	}
	b = append(b, valueEnc...)

	b = append(b, NL)

	//TODO int overflow int64
	ttlEnc := IntToBytes(int(ttl))

	b = append(b, ttlEnc...)
	return append(b, CR), nil
}
