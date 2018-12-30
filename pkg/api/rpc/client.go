package rpc

import (
	"context"
	"github.com/logrange/range/pkg/records"
)

type dataServiceStub struct {

}

func (dss *dataServiceStub) Write(src string, records records.Records) error {
	return nil
}

func (dss *dataServiceStub) Iterator(src string) (records.Records, error) {
	return nil, nil
}

type journalIteratorStub struct {
}

func (jis *journalIteratorStub) Close() error {
	return nil
}

func (jis *journalIteratorStub) Next(ctx context.Context) {

}

func (jis *journalIteratorStub) Get(ctx context.Context) (records.Record, error) {
	return nil, nil
}
