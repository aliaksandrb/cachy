package proto

import "errors"

var (
	ErrUnsupportedType = errors.New("unsupported type")
	ErrUnsupportedCmd  = errors.New("unsupported command")
	ErrBadMsg          = errors.New("malformed message")
	ErrBadDelimiter    = errors.New("bad delimiter")
	ErrUnknown         = errors.New("unknown error")
)
