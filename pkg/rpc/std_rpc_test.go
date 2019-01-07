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
