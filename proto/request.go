package proto

import "time"

type Req struct {
	Cmd   byte
	Key   string
	Value []byte
	TTL   time.Duration
}
