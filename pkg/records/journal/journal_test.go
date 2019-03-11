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

package journal

import (
	"github.com/logrange/range/pkg/records/chunk"
	"testing"
)

func BenchmarkJHashFromName(b *testing.B) {
	str := "some name for a journal"
	for i := 0; i < b.N; i++ {
		JHashFromName(str)
	}
}

func TestParsePos(t *testing.T) {
	testPos(t, Pos{})
	testPos(t, Pos{1341234, 123413})
	testPos(t, Pos{chunk.Id(0xFFFFFFFFFFFFFFFF), uint32(0xFFFFFFFF)})

	p1, err := ParsePos("")
	if err != nil || p1.CId != 0 || p1.Idx != 0 {
		t.Fatal("p1=", p1, " must be empty, err=", err)
	}
}

func testPos(t *testing.T, p Pos) {
	str := p.String()
	p1, err := ParsePos(str)
	if err != nil || p1 != p {
		t.Fatal("p1=", p1, " must be equal to ", p, ", err=", err)
	}
}
