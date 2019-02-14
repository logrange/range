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

package rpc

import (
	"fmt"
	"github.com/logrange/range/pkg/utils/encoding/xbinary"
	"io"
	"reflect"
	"testing"
)

type rwcs struct {
	buf   []byte
	idx   int
	toErr int
}

func (rwc *rwcs) Read(p []byte) (n int, err error) {
	n = copy(p, rwc.buf[rwc.idx:])
	rwc.idx += n
	if n == 0 && rwc.toErr != 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (rwc *rwcs) Write(p []byte) (n int, err error) {
	if rwc.toErr < 0 {
		return 0, fmt.Errorf("Write Error")
	}
	rwc.buf = append(rwc.buf, p...)
	if rwc.toErr > 0 {
		rwc.toErr -= len(p)
		if rwc.toErr == 0 {
			rwc.toErr = -1
		}
	}
	return len(p), nil
}

func (rwc *rwcs) Close() error {
	return nil
}

type testMsg []byte

func (tm testMsg) WritableSize() int {
	return len(tm)
}

func (tm testMsg) WriteTo(ow *xbinary.ObjectsWriter) (int, error) {
	return ow.WriteBytes(tm)
}

func TestClWriteCodec(t *testing.T) {
	rwc := &rwcs{[]byte{}, 0, 0}
	cc := newClntIOCodec(rwc)
	tm := testMsg([]byte{})
	cc.writeRequest(1, 2, tm)
	if !reflect.DeepEqual(rwc.buf, []byte{0, 0, 0, 1, 0, 2, 0, 0, 0, 0}) {
		t.Fatal("rwc.buf=", rwc.buf, " has unexpected value")
	}

	tm = testMsg([]byte{1, 2})
	rwc.buf = []byte{}
	cc.writeRequest(1, 2, tm)
	if !reflect.DeepEqual(rwc.buf, []byte{0, 0, 0, 1, 0, 2, 0, 0, 0, 2, 1, 2}) {
		t.Fatal("rwc.buf=", rwc.buf, " has unexpected value")
	}

	rwc.toErr = 5
	rwc.buf = []byte{}
	err := cc.writeRequest(1, 2, tm)
	if err != nil {
		t.Fatal("expecting err, but it is nil")
	}
}

func TestClReadResponse(t *testing.T) {
	rwc := &rwcs{[]byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 65}, 0, 0}
	cc := newClntIOCodec(rwc)
	reqId, opErr, bs, err := cc.readResponse()
	if reqId != 1 || opErr != nil || bs != 1 || err != nil {
		t.Fatal("Wrong data reqId=", reqId, ", opErr=", opErr, ", bz=", bs, ", err=", err)
	}
	buf := make([]byte, 1)
	err = cc.readResponseBody(buf)
	if err != nil || buf[0] != 65 {
		t.Fatal("Unexpected err=", err, ", or wrong buf val ", buf)
	}

	rwc.idx = 0
	rwc.buf[4] = 1
	reqId, opErr, bs, err = cc.readResponse()
	if reqId != 1 || opErr == nil || bs != 0 || err != nil || opErr.Error() != "A" {
		t.Fatal("Wrong data reqId=", reqId, ", opErr=", opErr, ", bz=", bs, ", err=", err)
	}

	kopErr := fmt.Errorf("AA")
	RegisterError(kopErr)
	rwc.buf = []byte{0, 0, 0, 1, 0, 1, 0, 0, 0, 2, 65, 65}
	rwc.idx = 0
	reqId, opErr, bs, err = cc.readResponse()
	if reqId != 1 || opErr != kopErr || bs != 0 || err != nil {
		t.Fatal("Wrong data reqId=", reqId, ", opErr=", opErr, ", bz=", bs, ", err=", err)
	}
}

func TestSrvReadRequest(t *testing.T) {
	rwc := &rwcs{[]byte{0, 0, 0, 1, 1, 0, 0, 0, 0, 1, 65}, 0, 0}
	sc := newSrvIOCodec("test", rwc)
	reqId, fnId, bs, err := sc.readRequest()
	if reqId != 1 || fnId != 256 || bs != 1 || err != nil {
		t.Fatal("Wrong data read reqId=", reqId, ", fnId=", fnId, ", bs=", bs, ", err=", err)
	}

	rwc.idx = 4
	rwc.toErr = 1
	_, _, _, err = sc.readRequest()
	if err != io.ErrUnexpectedEOF {
		t.Fatal("Must be ErrUnexpectedEOF, but err=", err)
	}
}

func TestSrvWriteResp(t *testing.T) {
	rwc := &rwcs{[]byte{}, 0, 0}
	sc := newSrvIOCodec("test", rwc)
	tm := testMsg([]byte{1, 2})
	sc.writeResponse(1, nil, tm)
	if !reflect.DeepEqual(rwc.buf, []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 2, 1, 2}) {
		t.Fatal("rwc.buf=", rwc.buf, " has unexpected value")
	}

	tm = testMsg([]byte{})
	rwc.buf = []byte{}
	sc.writeResponse(2, nil, tm)
	if !reflect.DeepEqual(rwc.buf, []byte{0, 0, 0, 2, 0, 0, 0, 0, 0, 0}) {
		t.Fatal("rwc.buf=", rwc.buf, " has unexpected value")
	}

	tm = testMsg([]byte{1, 2})
	rwc.buf = []byte{}
	err := sc.writeResponse(2, fmt.Errorf("AB"), tm)
	if err != nil || !reflect.DeepEqual(rwc.buf, []byte{0, 0, 0, 2, 0, 1, 0, 0, 0, 2, 65, 66}) {
		t.Fatal("rwc.buf=", rwc.buf, " has unexpected value, err=", err)
	}
}
