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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sync"
	"sync/atomic"
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
	res, err := mmf.Buffer(12345, len(buf))
	if err != nil {
		mmf.Close()
		t.Fatal("offs=12345 err=", err)
	}
	copy(res, buf)

	res, err = mmf.Buffer(mmf.Size()-2, len(buf))
	if len(res) != 2 || err != nil {
		mmf.Close()
		t.Fatal("offs=Size-2 Wrong n=", len(res), " which should be 2, or err=", err)
	}
	copy(res, buf)

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

	res, err = mmf.Buffer(12345, len(buf))
	if err != nil || len(res) != len(buf) || !reflect.DeepEqual(buf, res) {
		t.Fatal("Expected ", buf, ", but read ", res, ", n=", len(res), ", err=", err)
	}

	res, err = mmf.Buffer(mmf.Size()-3, len(buf))
	if err != nil || len(res) != 3 || !reflect.DeepEqual(buf[:2], res[1:3]) {
		t.Fatal("Expected ", buf[:2], ", but read ", res[1:3], ", n=", len(res), ", err=", err)
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
	res, err := mmf.Buffer(4093, len(buf))
	n := copy(res, buf)
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

	res, err = mmf.Buffer(4093, len(buf))
	if err != nil || len(res) != len(buf) || !reflect.DeepEqual(buf, res) {
		t.Fatal("Expected ", buf, ", but read ", res, ", n=", len(res), ", err=", err)
	}
	mmf.Close()
}

func TestParrallelMMFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestParrallelMMFile")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	fsz := int64(40960)
	fn := path.Join(dir, "testFile2")
	mmf, err := NewMMFile(fn, fsz)
	if err != nil {
		t.Fatal("Could not open the file ", fn, " first time err=", err)
	}
	defer mmf.Close()

	var wg sync.WaitGroup
	var errs int32

	for i := 0; i < 640; i++ {
		wg.Add(1)
		go func(pid int) {
			buf := make([]byte, 64)
			for j, _ := range buf {
				buf[j] = byte(pid)
			}
			for i := 0; i < 500; i++ {
				res, err := mmf.Buffer(int64(pid*64), len(buf))
				if len(res) != len(buf) || err != nil {
					fmt.Println("Error when write n=", len(res), ", err=", err)
					atomic.AddInt32(&errs, 1)
				}
				copy(res, buf)

				res, err = mmf.Buffer(int64(pid*64), len(buf))
				if len(res) != len(buf) || err != nil {
					fmt.Println("Error when read n=", len(res), ", err=", err)
					atomic.AddInt32(&errs, 1)
				}

				if !reflect.DeepEqual(buf, res) {
					fmt.Println("Error buf=", buf, ", res=", res)
					atomic.AddInt32(&errs, 1)
					break
				}
			}

			wg.Done()
		}(i)
	}
	wg.Wait()
	if atomic.LoadInt32(&errs) != 0 {
		t.Fatal(" errs=", atomic.LoadInt32(&errs))
	}
}
