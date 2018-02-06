package proto

// Supported commands.
const (
	CmdGet    = '#'
	CmdSet    = '+'
	CmdUpdate = '^'
	CmdRemove = '-'
	CmdKeys   = '~'
)

// Supported datatypes.
const (
	ERROR  = '!'
	STRING = '$'
	SLICE  = '@'
	MAP    = ':'
	NIL    = '*'
	INT    = '&'
)

// Escape chars.
const (
	NL = '\n'
	CR = '\r'
)

// Message types.
const (
	KindReq byte = iota
	KindRes
)
