package transport

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
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

	sCfg := Config{
		ListenAddr: "127.0.0.1:12345",
		TlsEnabled: boolPtr(false),
	}

	err := testConnectivity(sCfg, cCfg)
	if err != nil {
		t.Fatal(err)
	}
}

func Test2WayTlsConnectivity(t *testing.T) {
	sCfg := Config{
		ListenAddr:  "127.0.0.1:12345",
		TlsEnabled:  boolPtr(true),
		Tls2Way:     boolPtr(true),
		TlsCertFile: absCertPath("server.crt"),
		TlsKeyFile:  absCertPath("server.key"),
		TlsCAFile:   absCertPath("ca.crt"),
	}

	cCfg := Config{
		ListenAddr:  "127.0.0.1:12345",
		TlsEnabled:  boolPtr(true),
		TlsCertFile: absCertPath("client.crt"),
		TlsKeyFile:  absCertPath("client.key"),
		TlsCAFile:   absCertPath("ca.crt"),
	}

	err := testConnectivity(sCfg, cCfg)
	if err != nil {
		t.Fatal(err)
	}
}

func Test2WayTlsConnectivitySkipVerify(t *testing.T) {
	sCfg := Config{
		ListenAddr:  "127.0.0.1:12345",
		TlsEnabled:  boolPtr(true),
		Tls2Way:     boolPtr(true),
		TlsCertFile: absCertPath("server1.crt"),
		TlsKeyFile:  absCertPath("server1.key"),
		TlsCAFile:   absCertPath("ca.crt"),
	}

	cCfg := Config{
		ListenAddr:    "127.0.0.1:12345",
		TlsEnabled:    boolPtr(true),
		TlsSkipVerify: boolPtr(true),
		TlsCertFile:   absCertPath("client.crt"),
		TlsKeyFile:    absCertPath("client.key"),
		TlsCAFile:     absCertPath("ca.crt"),
	}

	err := testConnectivity(sCfg, cCfg)
	if err != nil {
		t.Fatal(err)
	}
}

func testConnectivity(sCfg Config, cCfg Config) error {
	sl, err := NewServerListener(sCfg)
	if err != nil {
		return errors.Wrap(err, "server listen error")
	}

	defer sl.Close()
	go runTestEchoSrv(sl)

	cc, err := NewClientConn(cCfg)
	if err != nil {
		return errors.Wrap(err, "client connect error")
	}

	defer cc.Close()
	msg := "Test connectivity!\n"
	n, err := cc.Write([]byte(msg))
	if err != nil {
		return errors.Wrap(err, "client write error")
	}

	buf := make([]byte, len(msg))
	n, err = cc.Read(buf)
	if err != nil || len(msg) != n || string(buf[:n]) != msg {
		return errors.Wrap(err, fmt.Sprint("client read error, err=", err, ", n=", n,
			", msg=", msg, ", resp=", string(buf[:n])))
	}
	return nil
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
