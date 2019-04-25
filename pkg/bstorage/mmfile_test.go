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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestOpenCloseMMFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestOpenCloseMMFile")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	fsz := int64(25000 * 4096)
	fn := path.Join(dir, "testFile")
	mmf, err := NewMMFile(fn, fsz)
	if err != nil {
		t.Fatal("Could not open the file ", fn, " first time err=", err)
	}

	buf := []byte{1, 2, 3, 4, 5}
	n, err := mmf.Write(12345, buf)
	if n != len(buf) || err != nil {
		mmf.Close()
		t.Fatal("offs=12345 Wrong n=", n, " which should be ", len(buf), ", or err=", err)
	}

	n, err = mmf.Write(mmf.Size()-2, buf)
	if n != 2 || err != nil {
		mmf.Close()
		t.Fatal("offs=Size-2 Wrong n=", n, " which should be 2, or err=", err)
	}

	mmf.Close()
	if mmf.Size() != -1 {
		t.Fatal("size must be -1, after closing")
	}

	mmf, err = NewMMFile(fn, -1)
	if err != nil {
		t.Fatal("Could not open the file ", fn, " second time err=", err)
	}
	defer mmf.Close()

	if mmf.Size() != fsz {
		t.Fatal("The mmapped file size must be ", fsz, ", but it is ", mmf.Size())
	}

	buf2 := make([]byte, 2*len(buf))
	n, err = mmf.Read(12345, buf2)
	if err != nil || n != len(buf2) || !reflect.DeepEqual(buf, buf2[:len(buf)]) {
		t.Fatal("Expected ", buf, ", but read ", buf2, ", n=", n, ", err=", err)
	}

	n, err = mmf.Read(mmf.Size()-3, buf2)
	if err != nil || n != 3 || !reflect.DeepEqual(buf[:2], buf2[1:3]) {
		t.Fatal("Expected ", buf[:2], ", but read ", buf2[1:3], ", n=", n, ", err=", err)
	}
}

func TestGrowMMFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestOpenCloseMMFile")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	fsz := int64(10 * 4096)
	fn := path.Join(dir, "testFile")
	mmf, err := NewMMFile(fn, fsz)
	if err != nil {
		t.Fatal("Could not open the file ", fn, " first time err=", err)
	}

	buf := []byte{1, 2, 3, 4, 5}
	n, err := mmf.Write(4093, buf)
	if n != len(buf) || err != nil {
		mmf.Close()
		t.Fatal("offs=4093 Wrong n=", n, " which should be ", len(buf), ", or err=", err)
	}

	err = mmf.Grow(4096)
	if err == nil {
		t.Fatal("err must be not nil, cause size is less than initial one")
	}

	err = mmf.Grow(20*4096 - 1)
	if err == nil {
		t.Fatal("err must be not nil, cause size is not divided by 4096")
	}

	if mmf.Size() != fsz {
		t.Fatal("The file still must be mapped")
	}

	err = mmf.Grow(20 * 4096)
	if err != nil {
		t.Fatal("expected err= nil, but it is ", err)
	}

	if mmf.Size() != 20*4096 {
		t.Fatal("The file size must be ", 20*4096, ", but it is ", mmf.Size())
	}

	buf2 := make([]byte, len(buf))
	n, err = mmf.Read(4093, buf2)
	if err != nil || n != len(buf2) || !reflect.DeepEqual(buf, buf2) {
		t.Fatal("Expected ", buf, ", but read ", buf2, ", n=", n, ", err=", err)
	}
	mmf.Close()
}
