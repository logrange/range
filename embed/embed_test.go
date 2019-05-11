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

package embed

import (
	"context"
	"github.com/logrange/range/pkg/records/journal"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jrivets/log4g"
	"github.com/logrange/range/pkg/records"
)

func TestSimple(t *testing.T) {
	dir, err := ioutil.TempDir("", "chunkTest")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir)
	log4g.SetLogLevel("", log4g.DEBUG)

	jc, err := NewJCtrlr(JCtrlrConfig{JournalsDir: dir})
	if err != nil {
		t.Fatal("Could not create new JCtrlr err=", err)
	}

	j, err := jc.GetOrCreate(nil, "test")
	if err != nil {
		t.Fatal("could not get new journal err=", err)
	}

	_, _, err = j.Write(context.Background(), records.SrtingsIterator("aaaa", "bbbb"))
	if err != nil {
		t.Fatal("Write err=", err)
	}
	j.Sync()
	it := journal.NewJIterator(j)
	r, _ := it.Get(context.Background())
	r = r.MakeCopy()
	it.Next(context.Background())
	r2, _ := it.Get(context.Background())
	if string(r) != "aaaa" || string(r2) != "bbbb" {
		t.Fatal("Unexpected ", string(r), " ", string(r2))
	}

	jc.Close()
}
