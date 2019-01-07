package rpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

type echoPacket []byte

func (ep echoPacket) EncodedSize() int {
	return len(ep)
}

func (ep echoPacket) Encode(writer io.Writer) error {
	return binary.Write(writer, binary.BigEndian, ep)
}

func echo(reqId int32, reqBody []byte, resp Responder) {
	resp.SendResponse(reqId, nil, echoPacket(reqBody))
}

func startEchoServer(port int) (Server, error) {
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

			scodec := NewSrvIOCodec(conn.RemoteAddr().String(), conn)
			err = srv.Serve(scodec)
			if err != nil {
				fmt.Println("Could not run Serve, err=", err)
			}
			fmt.Println("New connection for ", scodec)
		}
	}()
	return srv, nil
}

func newClient(port int) (Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	codec := NewClntIOCodec(conn)
	return NewClient(codec), nil
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

func TestOnce(t *testing.T) {
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

	end := time.Now().Add(time.Second)
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
