package proto

import "time"

type Request interface {
}

type Req struct {
	Cmd byte
	// TODO keep in bytes
	Key   string
	Value []byte
	TTL   time.Duration
}
