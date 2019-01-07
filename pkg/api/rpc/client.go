package rpc

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/logrange/range/pkg/records"
	"github.com/logrange/range/pkg/records/chunk"
	lrpc "github.com/logrange/range/pkg/rpc"
)

type clntEndpoint struct {
	client lrpc.Client
}

func (ce *clntEndpoint) Write(src string, records records.Records) error {
	ce.call(cFuncWrite, WritePacket{src, records})
	return nil
}

func (ce *clntEndpoint) GetChunkById() (chunk chunk.Chunk, err error) {

}

func (ce *clntEndpoint) call(funcId int32, p interface{}) {
	_, err := ce.client.Call(context.Background(), funcId, p.(proto.Message))
	if err != nil {

	}
}

type chunkStub struct {
}

type chunkInteratorStub struct {
}
