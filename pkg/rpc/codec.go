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
	"bufio"
	"encoding/binary"
	"github.com/logrange/range/pkg/utils/bytes"
	"io"
)

type (
	clntIOCodec struct {
		rwc  io.ReadWriteCloser
		wrtr *bufio.Writer
		hbuf [10]byte
	}

	srvIOCodec struct {
		id   string
		rwc  io.ReadWriteCloser
		wrtr *bufio.Writer
		hbuf [10]byte
	}
)

func newClntIOCodec(rwc io.ReadWriteCloser) *clntIOCodec {
	cc := new(clntIOCodec)
	cc.rwc = rwc
	cc.wrtr = bufio.NewWriter(rwc)
	return cc
}

func (cc *clntIOCodec) Close() error {
	cc.wrtr.Flush()
	return cc.rwc.Close()
}

func (cc *clntIOCodec) writeRequest(reqId int32, funcId int16, msg Encodable) error {
	err := binary.Write(cc.wrtr, binary.BigEndian, reqId)
	if err != nil {
		return err
	}

	err = binary.Write(cc.wrtr, binary.BigEndian, funcId)
	if err != nil {
		return err
	}

	sz := int32(msg.EncodedSize())
	err = binary.Write(cc.wrtr, binary.BigEndian, sz)
	if err != nil {
		return err
	}

	err = msg.Encode(cc.wrtr)

	cc.wrtr.Flush()
	return err
}

func (cc *clntIOCodec) readResponse() (reqId int32, opErr error, bodySize int, err error) {
	buf := cc.hbuf[:]
	_, err = io.ReadFull(cc.rwc, buf)
	if err != nil {
		return
	}
	reqId = int32(binary.BigEndian.Uint32(buf))
	errCode := int(binary.BigEndian.Uint16(buf[4:]))
	bodySize = int(binary.BigEndian.Uint32(buf[6:]))

	if errCode != 0 {
		buf = make([]byte, bodySize)
		bodySize = 0
		_, err = io.ReadFull(cc.rwc, buf)
		if err != nil {
			return
		}
		opErr = errorByText(bytes.ByteArrayToString(buf))
	}

	return
}

func (cc *clntIOCodec) readResponseBody(body []byte) error {
	_, err := io.ReadFull(cc.rwc, body)
	return err
}

func newSrvIOCodec(id string, rwc io.ReadWriteCloser) *srvIOCodec {
	sc := new(srvIOCodec)
	sc.id = id
	sc.rwc = rwc
	sc.wrtr = bufio.NewWriter(rwc)
	return sc
}

func (sc *srvIOCodec) Close() error {
	sc.wrtr.Flush()
	return sc.rwc.Close()
}

func (sc *srvIOCodec) readRequest() (reqId int32, funcId int16, bodySize int, err error) {
	buf := sc.hbuf[:]
	_, err = io.ReadFull(sc.rwc, buf)
	if err != nil {
		return
	}
	reqId = int32(binary.BigEndian.Uint32(buf))
	funcId = int16(binary.BigEndian.Uint16(buf[4:]))
	bodySize = int(binary.BigEndian.Uint32(buf[6:]))
	return
}

func (sc *srvIOCodec) readRequestBody(body []byte) error {
	_, err := io.ReadFull(sc.rwc, body)
	return err
}

func (sc *srvIOCodec) writeResponse(reqId int32, opErr error, msg Encodable) error {
	err := binary.Write(sc.wrtr, binary.BigEndian, reqId)
	if err != nil {
		return err
	}

	errCode := int16(0)
	if opErr != nil {
		errCode = int16(1)
	}

	err = binary.Write(sc.wrtr, binary.BigEndian, errCode)
	if err != nil {
		return err
	}

	if errCode != 0 {
		buf := bytes.StringToByteArray(opErr.Error())
		sz := int32(len(buf))
		err = binary.Write(sc.wrtr, binary.BigEndian, sz)
		if err != nil {
			return err
		}

		err = binary.Write(sc.wrtr, binary.BigEndian, buf)
		sc.wrtr.Flush()
		return err
	}

	sz := int32(msg.EncodedSize())
	err = binary.Write(sc.wrtr, binary.BigEndian, sz)
	if err != nil {
		return err
	}

	err = msg.Encode(sc.wrtr)
	sc.wrtr.Flush()
	return err
}
