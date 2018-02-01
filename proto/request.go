package proto

import "time"

type Request interface {
}

type Req struct {
	Cmd marker
	// TODO keep in bytes
	Key string
	// TODO keep in bytes
	Value interface{}
	TTL   time.Duration
}

type Res struct {
	Value interface{}
	// TODO keep in bytes
}
