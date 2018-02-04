package proto

type marker byte

// Command types.
const (
	CmdGet    marker = '#'
	CmdSet    marker = '+'
	CmdUpdate marker = '^'
	CmdRemove marker = '-'
	CmdKeys   marker = '~'
)

// Data types.
const (
	stringType marker = '$'
	sliceType  marker = '@'
	mapType    marker = ':'
	errType    marker = '!'
	nilType    marker = '*'
	intType    marker = '&'
)

const (
	ERROR  = '!'
	STRING = '$'
	SLICE  = '@'
	MAP    = ':'
	NIL    = '*'
	INT    = '&'
)

const (
	NL  = '\n'
	CR  = '\r'
	ESC = NL
)

const (
	nl byte = '\n'
	cr byte = '\r'
)

type msgKind byte

const (
	kindReq msgKind = iota
	kindRes
)
