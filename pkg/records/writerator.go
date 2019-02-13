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

package records

import (
	"context"
	"github.com/logrange/range/pkg/utils/encoding/xbinary"
)

// WrIterator a wrapper around an Iterator, which turns it to xbinary.Writable interface
type WrIterator struct {
	It Iterator
}

// Next is part of xbinary.Writable
func (wi *WrIterator) Next(ctx context.Context) {
	wi.It.Next(ctx)
}

// Get is part of xbinary.Writable
func (wi *WrIterator) Get(ctx context.Context) (xbinary.Writable, error) {
	return wi.It.Get(ctx)
}
