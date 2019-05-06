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
	"github.com/logrange/range/pkg/utils/errors"
	"reflect"
	"testing"
)

func TestPositioningInMem(t *testing.T) {
	ib := NewInMemBytes(100)
	buf := []byte{1, 2, 3, 4, 5}

	res, err := ib.Buffer(12, len(buf))
	if err != nil || len(res) != len(buf) {
		t.Fatal("offs=12 err=", err)
	}
	copy(res, buf)

	res, _ = ib.Buffer(13, len(buf)-1)
	if !reflect.DeepEqual(res, buf[1:]) {
		t.Fatal("must be properly copied")
	}

	res, err = ib.Buffer(98, 10)
	if err != nil || len(res) != 2 {
		t.Fatal("len(res)=", len(res), " or err=", err)
	}

	res, err = ib.Buffer(980, 10)
	if err == nil {
		t.Fatal("len(res)=", len(res), " and err = nil, but expected an error!")
	}
}

func TestCloseInMem(t *testing.T) {
	ib := NewInMemBytes(100)
	if err := ib.Close(); err != nil {
		t.Fatal("must not be errors here, but err=", err)
	}

	if ib.Size() != 0 {
		t.Fatal("the size must be 0 after close")
	}

	if err := ib.Close(); err != errors.ClosedState {
		t.Fatal("expected errors.ClosedState, but err=", err)
	}

}
