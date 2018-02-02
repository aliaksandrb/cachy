package proto

var (
	sep    = []byte{'\r', '\n'}
	nilEnc = []byte{byte(nilType), '\r', '\n'}
	//emptyStrEnc = []byte{'$', '\r', '\n'}
)
