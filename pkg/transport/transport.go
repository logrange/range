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

package transport

import (
	"fmt"
	"net"
)

type (
	Config struct {
		TlsEnabled  bool
		Tls2Way     bool
		TlsCAFile   string
		TlsKeyFile  string
		TlsCertFile string
		ListenAddr  string
	}
)

func (c Config) String() string {
	return fmt.Sprint(
		"\n\tTlsEnabled=", c.TlsEnabled,
		"\n\tTls2Way=", c.Tls2Way,
		"\n\tTlsCAFile=", c.TlsCAFile,
		"\n\tTlsKeyFile=", c.TlsKeyFile,
		"\n\tTlsCertFile=", c.TlsCertFile,
		"\n\tListenAddr=", c.ListenAddr,
	)
}

func NewServerListener(cfg Config) (net.Listener, error) {
	return net.Listen("tcp", cfg.ListenAddr)
}

func Dial(cfg Config) (net.Conn, error) {
	return net.Dial("tcp", cfg.ListenAddr)
}
