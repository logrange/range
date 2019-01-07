package rpc

type (
	WritePacket struct {
		Source  string
		Records []byte
	}
)
