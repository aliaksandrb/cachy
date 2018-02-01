package proto

import "time"

type Request struct {
	Cmd marker
	// TODO keep in bytes
	Key string
	// TODO keep in bytes
	Value interface{}
	TTL   time.Duration
}
