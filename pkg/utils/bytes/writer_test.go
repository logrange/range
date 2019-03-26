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

package bytes

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtendWriter(t *testing.T) {
	w := Writer{}
	w.Init(10, &Pool{})
	buf := []byte{1, 2, 3, 4, 5, 6, 7}
	w.Write(buf)
	if w.pos != len(buf) || len(w.buf) != 10 || !reflect.DeepEqual(buf, w.Buf()) {
		t.Fatal("something goes wrong ", w)
	}

	w.Write(buf)
	if w.pos != 2*len(buf) || len(w.buf) != 27 || !reflect.DeepEqual(buf, w.Buf()[len(buf):]) {
		t.Fatal("something goes wrong ", w, cap(w.buf))
	}
	w.Write(buf)
	w.Write(buf)
	if w.pos != 4*len(buf) || len(w.buf) != 61 || !reflect.DeepEqual(buf, w.Buf()[:len(buf)]) {
		t.Fatal("something goes wrong ", w)
	}

	ln := cap(w.buf)
	var bigData [1000]byte
	w.Write(bigData[:])
	if len(w.buf) != len(bigData)+2*ln || w.pos != 4*len(buf)+len(bigData) {
		t.Fatal("something goes wrong ", len(w.buf), " ", w.pos)
	}

	w.Close()
	if w.pool != nil || w.buf != nil {
		t.Fatal("must be cleaned up")
	}
}

func TestEmptyWriter(t *testing.T) {
	w := Writer{}
	buf := []byte{1, 2, 3, 4, 5, 6, 7}
	w.Write(buf)
	if w.pos != len(buf) || len(w.buf) != 74 || !reflect.DeepEqual(buf, w.Buf()) {
		t.Fatal("something goes wrong ", w)
	}

	w.WriteByte(1)
	if w.pos != len(buf)+1 || len(w.buf) != 74 {
		t.Fatal("something goes wrong ", w)
	}

	w.Reset()
	if w.pos != 0 || len(w.buf) != 74 {
		t.Fatal("buffer must not be touched")
	}
}

func TestInit10(t *testing.T) {
	w := Writer{}
	w.Init(10, nil)
	buf := []byte{1, 2, 3, 4, 5, 6, 7}
	w.Write(buf)
	if w.pos != len(buf) || len(w.buf) != 10 || !reflect.DeepEqual(buf, w.Buf()) {
		t.Fatal("something goes wrong ", w)
	}
	w.Write(buf)
	if w.pos != 2*len(buf) || len(w.buf) != 27 || !reflect.DeepEqual(buf, w.Buf()[len(buf):]) {
		t.Fatal("something goes wrong ", w, cap(w.buf))
	}

	w.WriteByte(12)
	if w.pos != 2*len(buf)+1 || w.buf[2*len(buf)] != 12 || len(w.buf) != 27 {
		t.Fatal("something goes wrong ", w)
	}

	w.Reset()
	if w.pos != 0 || len(w.buf) != 27 {
		t.Fatal("buffer must not be touched")
	}
}

func TestInit(t *testing.T) {
	str := "abc"
	w := Writer{}
	for i := 1; i < 10; i++ {
		w.Reset()
		var sb strings.Builder
		sz := 0
		for j := 1; j < i*100; j++ {
			sb.WriteString(str)
			w.WriteString(str)
			sz += len(str)
			if w.pos != sz || cap(w.buf) < sz {
				t.Fatal("something wrong")
			}
		}
		if sb.String() != ByteArrayToString(w.Buf()) {
			t.Fatal("expecting for i=", i, " ", sb.String(), " but got ", string(w.Buf()))
		}
	}
}
