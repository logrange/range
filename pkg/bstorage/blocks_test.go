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

package bstorage

import (
	"fmt"
	"github.com/logrange/range/pkg/utils/errors"
	"math/rand"
	"reflect"
	"testing"
)

func TestGetBlocksInSegment(t *testing.T) {
	if GetBlocksInSegment(1) != 9 || GetBlocksInSegment(16) != 129 {
		t.Fatal("something goes wrong GetBlocksInSegment(1)=", GetBlocksInSegment(1), ", GetBlocksInSegment(16)=", GetBlocksInSegment(16))
	}
	if GetBlocksInSegment(4096) != 32769 {
		t.Fatal("GetBlocksInSegment(4096)=", GetBlocksInSegment(4096))
	}
	if GetBlocksInSegment(6144) != -1 {
		t.Fatal("GetBlocksInSegment(6144)=", GetBlocksInSegment(6144))
	}
	if GetBlocksInSegment(8192) != 65537 {
		t.Fatal("GetBlocksInSegment(6144)=", GetBlocksInSegment(8192))
	}
	if GetBlocksInSegment(0) != -1 {
		t.Fatal("GetBlocksInSegment(0)=", GetBlocksInSegment(0))
	}
}

func TestNewBlocks(t *testing.T) {
	ib := NewInMemBytes(0)
	if _, err := NewBlocks(1024, ib, false); err == nil {
		t.Fatal("expecting err != nil here")
	}

	ib = NewInMemBytes(1024)
	// too small
	if _, err := NewBlocks(1024, ib, false); err == nil {
		t.Fatal("expecting err != nil here")
	}

	if _, err := NewBlocks(1, ib, true); err == nil {
		t.Fatal("expecting err != nil here")
	}

	if _, err := NewBlocks(1, ib, false); err != nil {
		t.Fatal("err=", err)
	}
}

func TestBlocksClose(t *testing.T) {
	ib := NewInMemBytes(1024)
	b, err := NewBlocks(1, ib, false)
	if err != nil {
		t.Fatal("expecting err is nil, but err=", err)
	}

	err = b.Close()
	if err != nil {
		t.Fatal("must be no errors, but err=", err)
	}

	if ib.Size() != 0 {
		t.Fatal("must be closed, but ib=", ib)
	}

	err = b.Close()
	if err != errors.ClosedState {
		t.Fatal("must be already closed, but err=", err)
	}
}

func TestBlockByIdx(t *testing.T) {
	// 1Mb
	ib := NewInMemBytes(1024 * 1024)
	for bs := 1; bs < 512; bs = bs << 1 {
		b, err := NewBlocks(bs, ib, false)
		if err != nil {
			t.Fatal("expecting err is nil, but err=", err)
		}
		maxIdx := colorBuffer(ib, bs)
		fmt.Println(maxIdx)
		for idx := 0; idx < maxIdx; idx++ {
			buf := makeBuf(idx, bs)
			buf1, err := b.Block(idx)
			if err != nil {
				t.Fatal("the block with idx=", idx, " could not be obtained, err=", err)
			}

			if !reflect.DeepEqual(buf, buf1) {
				t.Fatal("not equal, idx=", idx, ", bs=", bs, "\n buf=", buf, "\n buf1=", buf1)
			}
		}

		if _, err = b.Block(maxIdx); err == nil {
			t.Fatal("must be out of range, but err is nil for idx=", maxIdx)
		}
	}
}

func TestArrangeBlock(t *testing.T) {
	// 1Mb
	ib := NewInMemBytes(1024 * 1024)
	for bs := 1; bs < 512; bs = bs << 1 {
		buf, _ := ib.Buffer(0, int(ib.Size()))
		for i := 0; i < len(buf); i++ {
			buf[i] = 0
		}
		b, err := NewBlocks(bs, ib, false)
		if err != nil {
			t.Fatal("expecting err is nil, but err=", err)
		}

		if b.Available() != b.Count() {
			t.Fatal("total block count=", b.Count(), " but available ", b.Available(), " or err=", err)
		}

		maxIdx := colorBuffer(ib, bs)
		var idx, idx1 int
		for {
			idx1, err = b.ArrangeBlock()
			if err != nil {
				break
			}

			if idx != idx1 {
				t.Fatal("expecting idx=", idx, " but received idx=", idx1)
			}
			idx++
		}

		if b.Available() != 0 {
			t.Fatal("total block count=", b.Count(), " but available ", b.Available(), " or err=", err)
		}

		if maxIdx != idx {
			t.Fatal("should be ", maxIdx, " blocks allocated, but ", idx, " last err=", err)
		}
	}
}

func TestArrangeBlock2(t *testing.T) {
	// 1Mb
	ib := NewInMemBytes(1024 * 1024)
	for bs := 1; bs < 512; bs = bs << 1 {
		buf, _ := ib.Buffer(0, int(ib.Size()))
		for i := 0; i < len(buf); i++ {
			buf[i] = 0
		}
		b, err := NewBlocks(bs, ib, false)
		if err != nil {
			t.Fatal("expecting err is nil, but err=", err)
		}
		maxIdx := colorBuffer(ib, bs)
		for cnt := 0; cnt < maxIdx*3; cnt++ {
			idx1, err := b.ArrangeBlock()
			if err != nil {
				t.Fatal("must be able to arrange new block")
			}

			if err = b.FreeBlock(idx1 - 1); err == nil {
				t.Fatal("must be an error, but it is nil, for idx=", idx1-1)
			}
			if err = b.FreeBlock(idx1 + 1); err == nil {
				t.Fatal("must be an error, but it is nil, for idx=", idx1+1, "bs=", bs, ", cnt=", cnt)
			}
			if err = b.FreeBlock(idx1); err != nil {
				t.Fatal("must be no error, but it is ", err, ", for idx=", idx1)
			}
		}
	}
}

func TestArrangeBlock3(t *testing.T) {
	ib := NewInMemBytes(1024)
	for bs := 1; bs < 16; bs = bs << 1 {
		buf, _ := ib.Buffer(0, int(ib.Size()))
		for i := 0; i < len(buf); i++ {
			buf[i] = 0
		}
		b, err := NewBlocks(bs, ib, false)
		if err != nil {
			t.Fatal("expecting err is nil, but err=", err)
		}
		maxIdx := colorBuffer(ib, bs)
		for cnt := 0; cnt < maxIdx; cnt++ {
			_, err := b.ArrangeBlock()
			if err != nil {
				t.Fatal("must be able to arrange new block")
			}
		}

		for cnt := 0; cnt < maxIdx*2; cnt++ {
			idx := rand.Int() % maxIdx
			if err := b.FreeBlock(idx); err != nil {
				t.Fatal("Must be able to free idx=", idx, ", but err=", err)
			}

			if b.Available() != 1 {
				t.Fatal("total block count=", b.Count(), " but available ", b.Available(), ", and it must be 1, or err=", err)
			}

			_, err := b.ArrangeBlock()
			if err != nil {
				t.Fatal("must be able to arrange new block")
			}
		}
	}
}

func colorBuffer(ib Bytes, bs int) int {
	offs := 0
	idx := 0
	blks := GetBlocksInSegment(bs)
	for offs+blks*bs <= int(ib.Size()) {
		offs += bs
		for i := 0; i < blks-1; i++ {
			buf2, err := ib.Buffer(int64(offs), bs)
			if err != nil || len(buf2) != bs {
				panic(err)
			}
			copy(buf2, makeBuf(idx, bs))
			idx++
			offs += bs
		}
	}
	return idx
}

func makeBuf(idx, bs int) []byte {
	buf := make([]byte, bs)
	for j := 0; j < bs; j++ {
		buf[j] = byte(idx)
	}
	return buf
}
