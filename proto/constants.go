package proto

// Supported commands.
const (
	CmdGet    = '#'
	CmdSet    = '+'
	CmdUpdate = '^'
	CmdRemove = '-'
	CmdKeys   = '~'
	CmdValue  = '.'
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
	NL  = '\n'
	CR  = '\r'
	ESC = NL
)

// Message types.
const (
	KindReq byte = iota
	KindRes
)
