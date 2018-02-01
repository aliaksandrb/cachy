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
