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
	"context"
	"github.com/logrange/range/pkg/records"
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

func TestFormatId(t *testing.T) {
	s := Id(1).String()
	if "0000000000000001" != s {
		t.Fatal("Unexpected Id value=", s)
	}
}

func TestFindChunkById(t *testing.T) {
	testFindChunkById(t, Chunks{}, 123, -1)
	testFindChunkById(t, Chunks{&fakeChunk{1}}, 123, -1)
	testFindChunkById(t, Chunks{&fakeChunk{1123}}, 123, -1)
	testFindChunkById(t, Chunks{&fakeChunk{1}, &fakeChunk{2}, &fakeChunk{3}}, 1, 0)
	testFindChunkById(t, Chunks{&fakeChunk{1}, &fakeChunk{2}, &fakeChunk{3}}, 3, 2)
	testFindChunkById(t, Chunks{&fakeChunk{1}, &fakeChunk{2}, &fakeChunk{3}}, 4, -1)
	testFindChunkById(t, Chunks{&fakeChunk{10}, &fakeChunk{20}, &fakeChunk{30}}, 1, -1)
}

func testFindChunkById(t *testing.T, chks Chunks, cid Id, pos int) {
	res := FindChunkById(chks, cid)
	if res != pos {
		t.Fatal(" for chunks ", chks, " the resuls of search cid=", cid, " is ", res, ", but expected was ", pos)
	}
}

type fakeChunk struct {
	id Id
}

func (fc *fakeChunk) Close() error { panic("not supported"); return nil }
func (fc *fakeChunk) Id() Id       { return fc.id }
func (fc *fakeChunk) Write(ctx context.Context, it records.Iterator) (int, uint32, error) {
	panic("not supported")
	return 0, 0, nil
}
func (fc *fakeChunk) Sync()                       { panic("not supported") }
func (fc *fakeChunk) Iterator() (Iterator, error) { panic("not supported"); return nil, nil }
func (fc *fakeChunk) Size() int64                 { panic("not supported"); return 0 }
func (fc *fakeChunk) Count() uint32               { panic("not supported"); return 0 }
func (fc *fakeChunk) AddListener(lstnr Listener)  { panic("not supported") }
