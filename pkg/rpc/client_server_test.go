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
	"context"
	"fmt"
	"github.com/logrange/range/pkg/utils/encoding/xbinary"
	"github.com/logrange/range/pkg/utils/errors"
	"net"
	"reflect"
	"testing"
	"time"
)

type echoPacket []byte

func (ep echoPacket) WritableSize() int {
	return len(ep)
}

func (ep echoPacket) WriteTo(ow *xbinary.ObjectsWriter) (int, error) {
	return ow.WriteBytes(ep)
}

func echo(reqId int32, reqBody []byte, sc *ServerConn) {
	sc.SendResponse(reqId, nil, echoPacket(reqBody))
}

func noResp(reqId int32, reqBody []byte, sc *ServerConn) {

}

func tenMsDelay(reqId int32, reqBody []byte, sc *ServerConn) {
	time.Sleep(10 * time.Millisecond)
	sc.SendResponse(reqId, nil, echoPacket(reqBody))
}

func startEchoServer(port int) (*server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := NewServer()
	srv.Register(1, echo)

	go func() {
		defer srv.Close()
		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("error when accept err=", err)
				return
			}

			err = srv.Serve(conn.RemoteAddr().String(), conn)
			if err != nil {
				fmt.Println("Could not run Serve, err=", err)
			}
			fmt.Println("New connection for ", conn)
		}
	}()
	return srv, nil
}

func newClient(port int) (*client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	return NewClient(conn), nil
}

func BenchmarkEcho(b *testing.B) {
	srv, err := startEchoServer(1235)
	if err != nil {
		b.Fatal("Could not start server err=", err)
	}
	defer srv.Close()

	clnt, err := newClient(1235)
	if err != nil {
		b.Fatal("Could not start the client err=", err)
	}
	defer clnt.Close()

	msg := "testOnce"
	buf := []byte(msg)

	b.Run("bench", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res, opErr, err := clnt.Call(context.Background(), 1, echoPacket(buf))
			if opErr != nil || err != nil || len(res) != len(buf) {
				b.Fatal("Unexpected errors: opErr=", opErr, ", err=", err, " res=", string(res))
			}
		}
	})
}

func TestCountPerSecond(t *testing.T) {
	srv, err := startEchoServer(1234)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}
	defer srv.Close()

	clnt, err := newClient(1234)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}
	defer clnt.Close()

	msg := "testOnce"
	buf := []byte(msg)

	end := time.Now().Add(time.Millisecond)
	cnt := 0
	for time.Now().Before(end) {
		cnt++
		res, opErr, err := clnt.Call(context.Background(), 1, echoPacket(buf))
		if opErr != nil || err != nil || len(res) != len(buf) {
			t.Fatal("Unexpected errors: opErr=", opErr, ", err=", err)
		}
	}

	fmt.Println("Count ", cnt)
}

func TestClientClosed(t *testing.T) {
	srv, err := startEchoServer(1235)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}
	defer srv.Close()

	clnt, err := newClient(1235)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}

	res, opErr, err := clnt.Call(context.Background(), 1, echoPacket(nil))
	if opErr != nil || err != nil || len(res) != 0 {
		t.Fatal("Unexpected errors: opErr=", opErr, ", err=", err)
	}

	if len(srv.conns) != 1 {
		t.Fatal("expecting 1 connection is established")
	}

	clnt.Close()

	_, _, err = clnt.Call(context.Background(), 1, echoPacket(nil))
	if err != errors.ClosedState {
		t.Fatal("Unexpected error err=", err)
	}

	time.Sleep(5 * time.Millisecond)

	if len(srv.conns) != 0 {
		t.Fatal("expecting 0 connection is established")
	}
}

func TestServerClosed(t *testing.T) {
	srv, err := startEchoServer(1236)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}

	clnt, err := newClient(1236)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}

	res, opErr, err := clnt.Call(context.Background(), 1, echoPacket(nil))
	if opErr != nil || err != nil || len(res) != 0 {
		t.Fatal("Unexpected errors: opErr=", opErr, ", err=", err)
	}

	if clnt.isClosed() {
		t.Fatal("Expecting the client is not closed yet")
	}

	srv.Close()
	time.Sleep(10 * time.Millisecond)
	if !clnt.isClosed() {
		t.Fatal("Expecting the client closed")
	}

}

func TestServerLegClosed(t *testing.T) {
	srv, err := startEchoServer(1237)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}
	defer srv.Close()

	clnt, err := newClient(1237)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}

	if clnt.isClosed() {
		t.Fatal("Expecting the client is not closed yet")
	}

	for len(srv.conns) != 1 {
		time.Sleep(time.Millisecond)
	}

	for _, sc := range srv.conns {
		sc.Close()
	}
	time.Sleep(10 * time.Millisecond)
	if !clnt.isClosed() {
		t.Fatal("Expecting the client closed")
	}
}

func TestBlockedCallClosed(t *testing.T) {
	srv, err := startEchoServer(1238)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}

	clnt, err := newClient(1238)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}
	defer clnt.Close()

	srv.Register(2, noResp)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
	_, _, err = clnt.Call(ctx, 2, echoPacket(nil))
	if err != ctx.Err() {
		t.Fatal("Expecting err=", ctx.Err(), " but err=", err)
	}

	go func() {
		time.Sleep(time.Millisecond)
		srv.Close()
	}()

	_, _, err = clnt.Call(context.Background(), 2, echoPacket(nil))
	if err != errors.ClosedState {
		t.Fatal("Expecting err=ClosedState but err=", err)
	}
}

func TestLostResponse(t *testing.T) {
	srv, err := startEchoServer(1239)
	if err != nil {
		t.Fatal("Could not start server err=", err)
	}

	clnt, err := newClient(1239)
	if err != nil {
		t.Fatal("Could not start the client err=", err)
	}
	defer clnt.Close()

	srv.Register(3, tenMsDelay)

	msg := "echoPacket"
	buf := []byte(msg)

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Millisecond)
	res, opErr, err := clnt.Call(ctx, 3, echoPacket(buf))
	if opErr != nil || err != nil || len(res) != len(buf) {
		t.Fatal("Unexpected errors: opErr=", opErr, ", err=", err)
	}

	buf = []byte("lost")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Millisecond)
	_, _, err = clnt.Call(ctx, 3, echoPacket(buf))
	if err != ctx.Err() || err == nil {
		t.Fatal("Unexpected errors: ctx.Err=", ctx.Err(), ", err=", err)
	}

	if len(clnt.calls) != 0 {
		t.Fatal("Must be no responses waiting, but ", len(clnt.calls))
	}

	// waiting the server response
	time.Sleep(10 * time.Millisecond)

	buf = []byte("ok")
	res, _, err = clnt.Call(context.Background(), 3, echoPacket(buf))
	if err != nil || !reflect.DeepEqual(res, buf) {
		t.Fatal("Unexpected err=", err, " res=", res, ", buf=", buf)
	}
}
