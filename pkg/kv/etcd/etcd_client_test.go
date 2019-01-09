// Copyright 2018 The logrange Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/logrange/range/pkg/kv"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"
)

var nextKeyID int64

func nextKey() kv.Key {
	next := atomic.AddInt64(&nextKeyID, 1)
	return kv.Key(fmt.Sprintf("key-%d", next))
}

func TestMain(m *testing.M) {

	temp, err := ioutil.TempDir("/tmp", "etcd.test")
	if err != nil {
		panic(err)
	}

	cfg := embed.NewConfig()
	cfg.Dir = temp

	etcd, err := embed.StartEtcd(cfg)

	if err != nil {
		panic(err)
	}

	rt := m.Run()

	etcd.Close()

	<-etcd.Server.StopNotify()

	os.Exit(rt)
}

func storage() kv.Storage {
	st, err := New(&Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		panic(err)
	}
	return st
}

func TestCreate(t *testing.T) {

	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	v, err := s.Create(ctx, kv.Record{
		Key:   nextKey(),
		Value: []byte("data"),
	})

	assert.Nil(err)
	assert.Equal(v, kv.Version(1))
}

func TestRecreate(t *testing.T) {

	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	_, err = s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Equal(err, kv.ErrAlreadyExists)
}

func TestGet(t *testing.T) {

	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	rt, err := s.Get(ctx, key)
	assert.Nil(err)
	assert.Equal(rt.Version, ver(1))
	assert.Equal(rt.Value, val("data"))
}

func TestGetNotFound(t *testing.T) {

	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	rt, err := s.Get(ctx, key)
	assert.Equal(err, kv.ErrNotFound)
	assert.Equal(rt, empty())
}

func TestGetRange(t *testing.T) {

	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key1, key2 := nextKey(), nextKey()
	_, err := s.Create(ctx, kv.Record{
		Key:   key1,
		Value: []byte("data"),
	})
	assert.Nil(err)

	_, err = s.Create(ctx, kv.Record{
		Key:   key2,
		Value: []byte("data2"),
	})
	assert.Nil(err)

	rt, err := s.GetRange(ctx, key1, nextKey())
	assert.Nil(err)
	assert.Len(rt, 2)

	rt, err = s.GetRange(ctx, key1, nextKey())
	assert.Nil(err)
	assert.Len(rt, 2)
}
func TestLease(t *testing.T) {

	assert := assert.New(t)

	s := storage()

	ctx := context.Background()

	lease, err := s.Lessor().NewLease(ctx, time.Hour, true)
	assert.Nil(err)

	key := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key,
		Lease: lease.Id(),
		Value: val("data"),
	})
	assert.Nil(err)

	rt, err := s.Get(ctx, key)
	assert.Nil(err)
	assert.Equal(rt.Value, val("data"))
	assert.Equal(rt.Lease, lease.Id())

	err = lease.Release()
	assert.Nil(err)

	rt, err = s.Get(ctx, key)
	assert.Equal(err, kv.ErrNotFound)
}

func TestLeaseTimeout(t *testing.T) {

	assert := assert.New(t)

	s := storage()

	ctx := context.Background()

	lease, err := s.Lessor().NewLease(ctx, time.Second, false)
	assert.Nil(err)

	key := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key,
		Lease: lease.Id(),
		Value: val("data"),
	})
	assert.Nil(err)

	time.Sleep(4 * time.Second)

	_, err = s.Get(ctx, key)
	assert.Equal(kv.ErrNotFound, err)
}

func TestLeaseGet(t *testing.T) {

	assert := assert.New(t)

	s := storage()

	ctx := context.Background()

	lease, err := s.Lessor().NewLease(ctx, time.Hour, true)
	assert.Nil(err)

	key := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key,
		Lease: lease.Id(),
		Value: val("data"),
	})
	assert.Nil(err)

	rt, err := s.Get(ctx, key)
	assert.Nil(err)
	assert.Equal(rt.Value, val("data"))
	assert.Equal(rt.Lease, lease.Id())

	l, err := s.Lessor().GetLease(lease.Id())
	assert.Nil(err)

	err = l.Release()
	assert.Nil(err)

	rt, err = s.Get(ctx, key)
	assert.Equal(err, kv.ErrNotFound)
}

func TestLeaseMult(t *testing.T) {

	assert := assert.New(t)

	s := storage()

	ctx := context.Background()

	lease, err := s.Lessor().NewLease(ctx, time.Hour, true)
	assert.Nil(err)

	key1 := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key1,
		Lease: lease.Id(),
		Value: val("data"),
	})
	assert.Nil(err)

	key2 := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key2,
		Value: val("data"),
	})
	assert.Nil(err)

	key3 := nextKey()
	_, err = s.Create(ctx, kv.Record{
		Key:   key3,
		Lease: lease.Id(),
		Value: val("data"),
	})
	assert.Nil(err)

	assert.Nil(lease.Release())

	rt, err := s.GetRange(ctx, key1, nextKey())
	assert.Nil(err)
	assert.Len(rt, 1)
	assert.Equal(rt[0].Key, key2)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	err = s.Delete(ctx, key)
	assert.Nil(err)

	_, err = s.Get(ctx, key)
	assert.Equal(err, kv.ErrNotFound)
}

func TestCAS(t *testing.T) {
	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	rv, err := s.CasByVersion(ctx, kv.Record{
		Key:     key,
		Value:   val("new data"),
		Version: ver(1),
	})

	assert.Nil(err)
	assert.Equal(rv, ver(2))

	v, err := s.Get(ctx, key)
	assert.Nil(err)
	assert.Equal(v.Version, ver(2))
	assert.Equal(v.Value, val("new data"))
}

func TestCASFailed(t *testing.T) {
	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	_, err = s.CasByVersion(ctx, kv.Record{
		Key:     key,
		Value:   val("new data"),
		Version: ver(2),
	})

	assert.Equal(err, kv.ErrWrongVersion)

	v, err := s.Get(ctx, key)
	assert.Nil(err)
	assert.Equal(v.Version, ver(1))
	assert.Equal(v.Value, val("data"))
}

func TestWaitForVersion(t *testing.T) {
	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	retc := make(chan struct{})
	go func() {
		defer close(retc)
		r, err := s.WaitForVersionChange(ctx, key, ver(1))
		assert.Nil(err)
		assert.Equal(r.Version, ver(2))
		assert.Equal(r.Value, val("new data"))
	}()

	_, err = s.CasByVersion(ctx, kv.Record{
		Key:     key,
		Value:   val("new data"),
		Version: ver(1),
	})
	assert.Nil(err)

	select {
	case <-retc:
		return
	case <-time.After(10 * time.Second):
		assert.Fail("No result in 10 secs")
	}
}

func TestWaitForVersionDelete(t *testing.T) {
	assert := assert.New(t)

	s := storage()
	ctx := context.Background()

	key := nextKey()

	_, err := s.Create(ctx, kv.Record{
		Key:   key,
		Value: []byte("data"),
	})
	assert.Nil(err)

	retc := make(chan struct{})
	go func() {
		defer close(retc)
		r, err := s.WaitForVersionChange(ctx, key, ver(1))
		assert.Nil(err)
		assert.Equal(r, empty())
	}()

	time.Sleep(time.Second)
	err = s.Delete(ctx, key)
	assert.Nil(err)

	select {
	case <-retc:
		return
	case <-time.After(10 * time.Second):
		assert.Fail("No result in 10 secs")
	}
}

func empty() kv.Record {
	return kv.Record{}
}
func val(d string) kv.Value {
	return kv.Value(d)
}
func ver(v int) kv.Version {
	return kv.Version(v)
}
