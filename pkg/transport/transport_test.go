package transport

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"testing"
)

func TestNoTlsConnectivity(t *testing.T) {
	cCfg := Config{
		ListenAddr: "127.0.0.1:12345",
		TlsEnabled: boolPtr(false),
	}

	cc, err := NetClientConn(cCfg)
	if err == nil {
		t.Fatal("expected client error")
	}

	sCfg := Config{
		ListenAddr: "127.0.0.1:12345",
		TlsEnabled: boolPtr(false),
	}
	sl, err := NewServerListener(sCfg)
	if err != nil {
		t.Fatal("expected no server error=", err)
	}
	defer sl.Close()

	cc, err = NetClientConn(cCfg)
	if err != nil {
		t.Fatal("expected no client error=", err)
	}
	defer cc.Close()
}

func Test2WayTlsConnectivity(t *testing.T) {
	sCfg := Config{
		ListenAddr:  "127.0.0.1:12345",
		TlsEnabled:  boolPtr(true),
		Tls2Way: boolPtr(true),
		TlsCertFile: absCertPath("server0.crt"),
		TlsKeyFile:  absCertPath("server0.key"),
		TlsCAFile:   absCertPath("ca.pem"),
	}

	sl, err := NewServerListener(sCfg)
	if err != nil {
		t.Fatal("expected no server error=", err)
	}
	defer sl.Close()

	go runTestEchoSrv(sl)
	cCfg := Config{
		ListenAddr: "127.0.0.1:12345",
		TlsEnabled: boolPtr(true),
		TlsCertFile: absCertPath("client0.crt"),
		TlsKeyFile:  absCertPath("client0.key"),
		TlsCAFile:   absCertPath("ca.pem"),
	}

	cc, err := NetClientConn(cCfg)
	if err != nil {
		t.Fatal("expected no client error=", err)
	}
	defer cc.Close()

	msg := "Test TLS!\n"
	n, err := cc.Write([]byte(msg))
	if err != nil {
		t.Fatal("client write error=", err)
	}

	buf := make([]byte, len(msg))
	n, err = cc.Read(buf)
	if err != nil || len(msg) != n || string(buf[:n]) != msg {
		t.Fatal("client read error, err=", err, ", n=", n, ", msg=", msg, ", resp=", string(buf[:n]))
	}
}

func absCertPath(certFile string) string {
	absolute, err := filepath.Abs("./testdata/certs/" + certFile)
	if err != nil {
		panic(err)
	}
	return absolute
}

func runTestEchoSrv(sl net.Listener) {
	for {
		conn, err := sl.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			for {
				r := bufio.NewReader(conn)
				msg, err := r.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						fmt.Println("Error reading from conn, err=", err)
					}
					return
				}

				n, err := conn.Write([]byte(msg))
				if err != nil || n != len(msg) {
					fmt.Println("Error writing to conn, err=", err)
					return
				}
			}
		}(conn)
	}
}