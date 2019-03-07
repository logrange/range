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
	"testing"
)

func TestExtendWriter(t *testing.T) {
	w := Writer{}
	w.Reset(10, &Pool{})
	buf := []byte{1, 2, 3, 4, 5, 6, 7}
	w.Write(buf)
	if w.pos != len(buf) || len(w.buf) != 10 || !reflect.DeepEqual(buf, w.Buf()) {
		t.Fatal("something goes wrong ", w)
	}

	w.Write(buf)
	if w.pos != 2*len(buf) || len(w.buf) != 18 || !reflect.DeepEqual(buf, w.Buf()[len(buf):]) {
		t.Fatal("something goes wrong ", w)
	}
	w.Write(buf)
	if w.pos != 3*len(buf) || len(w.buf) != 33 || !reflect.DeepEqual(buf, w.Buf()[:len(buf)]) {
		t.Fatal("something goes wrong ", w)
	}

	ln := len(w.buf)
	var bigData [1000]byte
	w.Write(bigData[:])
	if len(w.buf) != 2*len(bigData)+ln || w.pos != 3*len(buf)+len(bigData) {
		t.Fatal("something goes wrong ", len(w.buf), " ", w.pos)
	}

	w.Close()
	if w.pool != nil || w.buf != nil {
		t.Fatal("must be cleaned up")
	}
}
