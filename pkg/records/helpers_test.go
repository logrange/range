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

package records

import (
	"io"
	"testing"
)

func TestSrtingsIterator(t *testing.T) {
	it := SrtingsIterator()
	if _, err := it.Get(nil); err != io.EOF {
		t.Fatal("should be EOF")
	}

	strs := []string{"aaa", "bbb", "ccc"}
	it = SrtingsIterator(strs...)
	cnt := 0
	for {
		rec, err := it.Get(nil)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal("Unexpected error ", err)
		}

		if string(rec) != strs[cnt] {
			t.Fatal("Expected ", strs[cnt], " but got ", string(rec))
		}
		cnt++
		it.Next(nil)
	}

	if cnt != len(strs) {
		t.Fatal("cnt must be ", len(strs), " but it is ", cnt)
	}

}

func TestByteArrayQuickCasts(t *testing.T) {
	b := make([]byte, 10)
	s := ByteArrayToString(b)
	b[0] = 'a'
	if len(s) != 10 || s[0] != 'a' {
		t.Fatal("must be pointed to same object s=", s, " b=", b)
	}

	s = "Hello WOrld"
	bf := StringToByteArray(string(s))
	s2 := ByteArrayToString(bf)
	if s != s2 {
		t.Fatal("Oops, expecting s1=", s, ", but really s2=", s2)
	}

	bf = StringToByteArray("")
	s2 = ByteArrayToString(bf)
	if s2 != "" {
		t.Fatal("Oops, expecting empty string, but got s2=", s2)
	}
}
