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
	"context"
	"fmt"
	"github.com/logrange/range/pkg/records"
	"sort"

	"github.com/jrivets/log4g"
	"github.com/logrange/range/pkg/records/chunk"
	"github.com/logrange/range/pkg/utils/errors"
)

type (
	journal struct {
		cc     ChnksController
		logger log4g.Logger
	}
)

// New creates a new journal
func New(cc ChnksController) Journal {
	j := new(journal)
	j.cc = cc
	j.logger = log4g.GetLogger("journal").WithId("{" + j.cc.JournalName() + "}").(log4g.Logger)
	j.logger.Info("New instance created ", j)
	return j
}

// Name returns the name of the journal
func (j *journal) Name() string {
	return j.cc.JournalName()
}

// Write - writes records received from the iterator to the journal. The Write can write only a portion from
// the it. It can happen when max chunk size is hit.
func (j *journal) Write(ctx context.Context, rit records.Iterator) (int, Pos, error) {
	var err error
	for _, err = rit.Get(ctx); err == nil; {
		c, err := j.cc.GetChunkForWrite(ctx)
		if err != nil {
			return 0, Pos{}, err
		}

		// If c.Write has an error, it will return n and offs actually written. So if n > 0 let's
		// consider the operation successful
		n, offs, err := c.Write(ctx, rit)
		if n > 0 {
			return n, Pos{c.Id(), offs}, nil
		}

		if err != errors.MaxSizeReached {
			break
		}
	}
	return 0, Pos{}, err
}

// Sync tells write chunk to be synced. Only known application for it is in tests
func (j *journal) Sync() {
	c, _ := j.cc.GetChunkForWrite(context.Background())
	if c != nil {
		c.Sync()
	}
}

func (j *journal) Iterator() Iterator {
	return &iterator{j: j}
}

func (j *journal) Size() uint64 {
	chunks, _ := j.cc.Chunks(nil)
	var sz int64
	for _, c := range chunks {
		sz += c.Size()
	}
	return uint64(sz)
}

func (j *journal) Count() uint64 {
	chunks, _ := j.cc.Chunks(nil)
	var cnt uint64
	for _, c := range chunks {
		cnt += uint64(c.Count())
	}
	return cnt
}

func (j *journal) String() string {
	return fmt.Sprintf("{name=%s}", j.cc.JournalName())
}

// getChunkById is looking for a chunk with cid. If there is no such chunk
// in the list, it will return the chunk with lowest Id, but which is greater
// then the cid. If there is no such chunks, so cid points is out
// of the chunks range, then the method returns nil
func (j *journal) getChunkById(cid chunk.Id) chunk.Chunk {
	chunks, _ := j.cc.Chunks(context.Background())
	n := len(chunks)
	if n == 0 {
		return nil
	}

	idx := sort.Search(len(chunks), func(i int) bool { return chunks[i].Id() >= cid })
	// according to the condition idx is always in [0..n]
	if idx < n {
		return chunks[idx]
	}
	return nil
}

// Chunks please see journal.Journal interface
func (j *journal) Chunks() ChnksController {
	return j.cc
}
