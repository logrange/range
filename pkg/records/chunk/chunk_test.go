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

package chunk

import (
	"testing"
	"time"
)

func TestNewCId(t *testing.T) {
	lastCid = 0
	cid := NewId()
	cid2 := NewId()
	diff := (cid2 - cid) >> 16
	if cid == cid2 || cid2 < cid || diff > 1 {
		t.Fatal("Ooops something wrong with cid generation cid=", cid, " cid2=", cid2, ", diff=", diff)
	}

	// some diff
	cid = NewId()
	time.Sleep(time.Millisecond)
	cid2 = NewId()
	diff = (cid2 - cid) >> 16
	if cid == cid2 || cid2 < cid || diff > 10000 || cid2 != Id(lastCid) {
		t.Fatal("Ooops something wrong with cid generation cid=", cid, " cid2=", cid2, ", diff=", diff)
	}

	// now put far to the future
	lastCid += uint64(time.Hour)
	lcid := Id(lastCid)
	cid = NewId()
	cid2 = NewId()
	diff = (cid - lcid) >> 16
	diff2 := (cid2 - lcid) >> 16
	if diff != 1 || diff2 != 2 {
		t.Fatal("expecting diff1=1 and diff2=2, but got ", diff, " and ", diff2, " respectively")
	}
}

func TestParsId(t *testing.T) {
	if _, err := ParseId("bbbs"); err == nil {
		t.Fatal("Expecting an error, but got nil")
	}

	if _, err := ParseId("-2"); err == nil {
		t.Fatal("Expecting an error, but got nil")
	}

	if _, err := ParseId("ffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("Expecting an error, but got nil")
	}

	if _, err := ParseId("ff"); err != nil {
		t.Fatal("Expecting no error, but got err=", err)
	}
}

func TestFormatßId(t *testing.T) {
	s := Id(1).String()
	if "0000000000000001" != s {
		t.Fatal("Unexpected Id value=", s)
	}
}
