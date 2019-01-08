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

package rpc

import (
	"fmt"
	"net"
	srpc "net/rpc"
	"testing"
	"time"
)

type TestEcho struct {
}

func (te *TestEcho) Fn(in string, out *string) error {
	*out = in
	return nil
}

// testStandartRpc counts number of calls per second
func testStandartRpc(t *testing.T) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", 1234))
	if err != nil {
		t.Fatal("Could not start to listen")
	}

	srv := srpc.NewServer()
	defer ln.Close()
	err = srv.Register(new(TestEcho))
	if err != nil {
		t.Fatal("Could not register ", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			srv.ServeConn(conn)
		}
	}()

	client, err := srpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fatal("could not dial ", err)
	}

	end := time.Now().Add(time.Second)
	cnt := 0
	for time.Now().Before(end) {
		cnt++
		var s string
		err = client.Call("TestEcho.Fn", "testOnce", &s)
		if err != nil || s != "testOnce" {
			t.Fatal("expecting testOnce, but ", s, ", or err=", err)
		}
	}
	fmt.Println("Count ", cnt)
}
