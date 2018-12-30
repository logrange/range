package rpc

import (
	"io"
)

type (
	Service struct {
	}

	Message interface {
		Size() int32
		Marshal(writer io.Writer) error
	}

	ClientConn interface {
		Call(ObjId int64, funcId int32, m Message) (ServerResponse, error)
	}

	ServerRespFunc func(resp ServerResponse)

	ServerConn interface {
		ServerResponse(clientId int32, srvResp ServerResponse) error
	}

	ServerFunc func(request ServerRequest) (done bool)

	ServerRequest struct {
		ClientId int32
		ObjId    int64
		Conn     ServerConn
		Data     []byte
	}

	ServerResponse struct {
		Error          error
		EncodedMessage []byte
	}
)
