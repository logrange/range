package rpc

import (
	"github.com/logrange/range/pkg/api"
	"io"
)

const (
	cFuncWrite = 1
)

func NewClient(addr string) (client api.Endpoint, err error) {

}

func NewServer(lstnOn string, ep api.Endpoint) (srvEndp io.Closer, err error) {

}
