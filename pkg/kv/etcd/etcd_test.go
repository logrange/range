// Copyright 2018-2019 The logrange Authors
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

/*
kv package contains interfaces and structures for working with a key-value storage.
The kv.Storage can be implemented as a distributed consistent storage like etcd,
consul etc. For stand-alone or test environments some light-weight implementations
like local file-storage or in-memory implementations could be used.
*/

package etcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/logrange/range/pkg/kv"
)

func stopStorage(s kv.Storage) {
	etcd := s.(*embeddedStorage).server
	etcd.Close()
	<-etcd.Server.StopNotify()
}

func tempDir() string {
	temp, err := ioutil.TempDir("", "etcd.test")
	if err != nil {
		panic(fmt.Errorf("Couldn't create temp dir : %v", err))
	}
	return temp
}

func testConfig() *Config {
	return &Config{
		Embedded:        true,
		EmbedDir:        tempDir(),
		EmbedListenURLs: []string{"http://localhost:18888"},
		EmbedClientURLs: []string{"http://localhost:18886"},
	}
}

func TestEmbedEtcdPut(t *testing.T) {

	st, err := New(testConfig())
	if err != nil {
		t.Fatal(err)
	}

	defer stopStorage(st)

	ctx := context.Background()

	v, err := st.Create(ctx, kv.Record{
		Key:   "foo",
		Value: []byte("bar"),
	})
	if err != nil {
		t.Error(err)
	}
	if v != 1 {
		t.Errorf("Invalid version: %q instead of %q", v, 1)
	}
}

func TestEmbedEtcdPut2(t *testing.T) {
	st, err := New(testConfig())
	if err != nil {
		t.Fatal(err)
	}

	defer stopStorage(st)

	ctx := context.Background()

	_, err = st.Create(ctx, kv.Record{
		Key:   "foo",
		Value: []byte("bar"),
	})
	if err != nil {
		t.Error(err)
	}

	_, err = st.Create(ctx, kv.Record{
		Key:   "foo",
		Value: []byte("buz"),
	})
	if err != kv.ErrAlreadyExists {
		t.Errorf("Error expected here: %v", kv.ErrAlreadyExists)
	}
}

func TestEtcd(t *testing.T) {
	cfg := Config{
		Endpoints: []string{"127.0.0.1:2379"},
	}
	cl, err := New(&cfg)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	rec := kv.Record{
		Key:   "test-111",
		Value: kv.Value("boo"),
	}
	v, err := cl.Create(ctx, rec)
	if err != nil {
		t.Error(err)
	}
	v, err = cl.Create(ctx, rec)
	if err == nil {
		t.Error(err)
	}

	rec.Value = kv.Value("ff")
	v, err = cl.Create(ctx, rec)
	if err == nil {
		t.Error(err)
	}
	_ = v
}
