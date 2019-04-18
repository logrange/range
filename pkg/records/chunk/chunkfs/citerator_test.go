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

package chunkfs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/logrange/range/pkg/records"
	"github.com/logrange/range/pkg/utils/fileutil"
)

func TestCIteratorCommon(t *testing.T) {
	// prepare the staff
	dir, err := ioutil.TempDir("", "citeratorCommonTest")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	p := NewFdPool(2)
	defer p.Close()

	fn := path.Join(dir, "tst")
	cw := newCWriter(fn, 0, 1000, 0)
	cw.ensureFWriter() // to create the file
	defer cw.Close()

	p.register(123, frParams{fname: fn, bufSize: ChnkIdxReaderBufSize})
	p.register(124, frParams{fname: fileutil.SetFileExt(fn, ChnkIndexExt), bufSize: ChnkIdxReaderBufSize})
	buf := make([]byte, 10)

	// test it now
	ci := newCIterator(123, p, &cw.cntCfrmd, buf)

	// empty it
	_, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}
	ci.Next(context.Background())

	si := records.SrtingsIterator("aa")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	res, err := ci.Get(context.Background())
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	res, err = ci.Get(context.Background())
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	ci.Release()
	res, err = ci.Get(context.Background())
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	ci.Next(context.Background())
	ci.Next(context.Background())
	ci.Next(context.Background())
	_, err = ci.Get(context.Background())
	if err != io.EOF || ci.Pos() != 1 {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	// wrong offset
	ci.SetPos(2)
	_, err = ci.Get(context.Background())
	if err != io.EOF || ci.Pos() != 1 {
		t.Fatal("Expecting ErrCorruptedData, but got err=", err, " or pos=", ci.Pos(), " is not 1")
	}

	si = records.SrtingsIterator("bb")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	ci.SetPos(0)
	_, err = ci.Get(context.Background())
	if err != nil {
		t.Fatal("Expecting nil but got err=", err)
	}

	ci.Next(context.Background())
	res, err = ci.Get(context.Background())
	if string(res) != "bb" || err != nil {
		t.Fatal("expecting bb, but got ", string(res), " and err=", err)
	}

	ci.Next(context.Background())
	_, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	// wrong offset
	ci.SetPos(12345)
	_, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	ci.SetPos(0)
	ci.Next(context.Background())
	res, err = ci.Get(context.Background())
	if records.ByteArrayToString(res) != "bb" || err != nil {
		t.Fatal("expecting bb, but got ", records.ByteArrayToString(res), " and err=", err)
	}
}

func TestCIteratorPos(t *testing.T) {
	// prepare the staff
	dir, err := ioutil.TempDir("", "citeratorPosTest")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	p := NewFdPool(2)
	defer p.Close()

	fn := path.Join(dir, "tst")
	cw := newCWriter(fn, 0, 1000, 0)
	cw.ensureFWriter() // to create the file
	defer cw.Close()

	p.register(123, frParams{fname: fn, bufSize: ChnkIdxReaderBufSize})
	p.register(124, frParams{fname: fileutil.SetFileExt(fn, ChnkIndexExt), bufSize: ChnkIdxReaderBufSize})
	buf := make([]byte, 10)

	// test it now
	ci := newCIterator(123, p, &cw.cntCfrmd, buf)

	si := records.SrtingsIterator("aa")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	res, err := ci.Get(context.Background())
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	ci.SetPos(0)
	res, err = ci.Get(context.Background())
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	// wrong offset
	ci.SetPos(5)
	res, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err, "res=", string(res), " pos=", ci.Pos())
	}

	ci.SetPos(1235)
	_, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	si = records.SrtingsIterator("bb")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	res, err = ci.Get(context.Background())
	if string(res) != "bb" || err != nil {
		t.Fatal("expecting bb, but got ", string(res), " and err=", err)
	}

	ci.Next(context.Background())
	_, err = ci.Get(context.Background())
	if err != io.EOF {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}
}

func TestCIteratorBackAndForth(t *testing.T) {
	// prepare the staff
	dir, err := ioutil.TempDir("", "citeratorBackAndForthTest")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	p := NewFdPool(2)
	defer p.Close()

	fn := path.Join(dir, "tst")
	cw := newCWriter(fn, 0, 1000, 0)
	cw.ensureFWriter() // to create the file
	defer cw.Close()

	p.register(123, frParams{fname: fn, bufSize: ChnkIdxReaderBufSize})
	p.register(124, frParams{fname: fileutil.SetFileExt(fn, ChnkIndexExt), bufSize: ChnkIdxReaderBufSize})
	buf := make([]byte, 10)

	// test it now
	ci := newCIterator(123, p, &cw.cntCfrmd, buf)

	si := records.SrtingsIterator("aa", "bb")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	// forth
	ctx := context.Background()
	ci.SetPos(0)
	for i := 0; i < 3; i++ {
		res, err := ci.Get(ctx)
		if string(res) != "aa" || err != nil {
			t.Fatal("expecting aa, but got ", string(res), " and err=", err)
		}
		ci.Next(ctx)
		ci.Next(ctx)
		_, err = ci.Get(ctx)
		if err != io.EOF {
			t.Fatal("Expecting io.EOF, but got err=", err)
		}

		ci.SetPos(1)
		res, err = ci.Get(ctx)
		if string(res) != "bb" || err != nil {
			t.Fatal("expecting bb, but got ", string(res), " and err=", err)
		}
		ci.Release()
		res, err = ci.Get(ctx)
		if string(res) != "bb" || err != nil {
			t.Fatal("expecting bb, but got ", string(res), " and err=", err)
		}
		ci.Release()
		ci.Next(ctx)
		_, err = ci.Get(ctx)
		if err != io.EOF {
			t.Fatal("Expecting io.EOF, but got err=", err)
		}
		ci.SetPos(0)
	}
}

func TestCIteratorBackAndForth2(t *testing.T) {
	// prepare the staff
	dir, err := ioutil.TempDir("", "citeratorBackAndForthTest2")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	p := NewFdPool(2)
	defer p.Close()

	fn := path.Join(dir, "tst")
	cw := newCWriter(fn, 0, 100000, 0)
	cw.ensureFWriter() // to create the file
	defer cw.Close()

	p.register(123, frParams{fname: fn, bufSize: ChnkIdxReaderBufSize})
	p.register(124, frParams{fname: fileutil.SetFileExt(fn, ChnkIndexExt), bufSize: ChnkIdxReaderBufSize})
	buf := make([]byte, 10)

	// test it now
	ci := newCIterator(123, p, &cw.cntCfrmd, buf)

	msgs := make([]string, 10000)
	for i := 0; i < len(msgs); i++ {
		msgs[i] = fmt.Sprintf("%d", i)
	}

	si := records.SrtingsIterator(msgs...)
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	// forth
	ctx := context.Background()
	ci.SetPos(-1)
	bkwrd := false
	idx := 0
	for i := 0; i < 10; i++ {
		ci.SetBackward(bkwrd)
		for {
			res, err := ci.Get(ctx)
			if err != nil {
				if err != io.EOF {
					t.Fatal("Expectin io.EOF, but err=", err)
				}

				break
			}

			if string(res) != msgs[idx] {
				t.Fatal("expecting ", msgs[idx], " at idx=", idx, " but got ", string(res))
			}

			if bkwrd {
				idx--
			} else {
				idx++
			}
			ci.Next(ctx)
		}
		if idx != -1 && idx != len(msgs) {
			t.Fatal("idx must be either 0, or ", len(msgs), ", but it is ", idx)
		}

		bkwrd = !bkwrd
		if bkwrd {
			idx = len(msgs) - 1
		} else {
			idx = 0
		}

	}
}

func TestCIteratorEmpty(t *testing.T) {
	// prepare the staff
	dir, err := ioutil.TempDir("", "citeratorBackAndForthTest")
	if err != nil {
		t.Fatal("Could not create new dir err=", err)
	}
	defer os.RemoveAll(dir) // clean up

	p := NewFdPool(2)
	defer p.Close()

	fn := path.Join(dir, "tst")
	cw := newCWriter(fn, 0, 1000, 0)
	cw.ensureFWriter() // to create the file
	defer cw.Close()

	p.register(123, frParams{fname: fn, bufSize: ChnkIdxReaderBufSize})
	p.register(124, frParams{fname: fileutil.SetFileExt(fn, ChnkIndexExt), bufSize: ChnkIdxReaderBufSize})
	buf := make([]byte, 10)

	// test it now
	ctx := context.Background()
	ci := newCIterator(123, p, &cw.cntCfrmd, buf)
	_, err = ci.Get(ctx)
	if err != io.EOF || ci.pos != 0 {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	ci.Next(ctx)
	_, err = ci.Get(ctx)
	if err != io.EOF || ci.pos != 0 {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	// backward
	ci.SetBackward(true)
	_, err = ci.Get(ctx)
	if err != io.EOF || ci.pos != -1 {
		t.Fatal("Expecting io.EOF, but got err=", err, " ci.pos=", ci.pos)
	}
	ci.Next(ctx)
	_, err = ci.Get(ctx)
	if err != io.EOF || ci.pos != -1 {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	ci.SetBackward(false)

	si := records.SrtingsIterator("aa", "bb")
	_, _, err = cw.write(nil, si)
	if err != nil {
		t.Fatal("could not write data to file ", fn, ", err=", err)
	}
	cw.flush()

	// forward
	res, err := ci.Get(ctx)
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	ci.SetBackward(true)
	res, err = ci.Get(ctx)
	if string(res) != "aa" || err != nil {
		t.Fatal("expecting aa, but got ", string(res), " and err=", err)
	}

	ci.Next(ctx)
	_, err = ci.Get(ctx)
	if err != io.EOF || ci.pos != -1 {
		t.Fatal("Expecting io.EOF, but got err=", err)
	}

	ci.Close()
}
