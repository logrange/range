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
package fileutil

import (
	"testing"
)

func TestSetFileExt(t *testing.T) {
	if s := SetFileExt("/a/b/c.ddd", ".idx"); "/a/b/c.idx" != s {
		t.Fatal("expecting \"/a/b/c.idx\" but got ", s)
	}

	if s := SetFileExt("/a/b/c.ddd", "idx"); "/a/b/c.idx" != s {
		t.Fatal("expecting \"/a/b/c.idx\" but got ", s)
	}
	if s := SetFileExt("abcd.txt", ""); "abcd" != s {
		t.Fatal("expecting \"abcd\" but got ", s)
	}
}

func TestEscapeUnescapeFileName(t *testing.T) {
	tname := "te*sst/bb;b\\ll:l\\0;:te||#03'()%@&\""
	res := UnescapeFileName(EscapeToFileName(tname))
	if tname != res {
		t.Fatal("Expecting same name as tname=", tname, " but got ", res)
	}
}
