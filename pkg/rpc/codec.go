package rpc

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

func NewClntIOCodec(rwc io.ReadWriteCloser) *clntIOCodec {
	cc := new(clntIOCodec)
	cc.rwc = rwc
	cc.wrtr = bufio.NewWriter(rwc)
	return cc
}

func (cc *clntIOCodec) Close() error {
	cc.wrtr.Flush()
	return cc.rwc.Close()
}

func (cc *clntIOCodec) WriteRequest(reqId int32, funcId int16, msg Encodable) error {
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

func (cc *clntIOCodec) ReadResponse() (reqId int32, opErr error, bodySize int, err error) {
	buf := cc.hbuf[:]
	_, err = io.ReadFull(cc.rwc, buf)
	if err != nil {
		return
	}
	reqId = int32(binary.BigEndian.Uint32(buf))
	errCode := int(binary.BigEndian.Uint16(buf[4:]))

	if errCode != 0 {
		buf = make([]byte, bodySize)
		_, err = io.ReadFull(cc.rwc, buf)
		if err != nil {
			return
		}
		opErr = fmt.Errorf(string(buf))
	} else {
		bodySize = int(binary.BigEndian.Uint32(buf[6:]))
	}

	return
}

func (cc *clntIOCodec) ReadResponseBody(body []byte) error {
	_, err := io.ReadFull(cc.rwc, body)
	return err
}

func NewSrvIOCodec(id string, rwc io.ReadWriteCloser) *srvIOCodec {
	sc := new(srvIOCodec)
	sc.id = id
	sc.rwc = rwc
	sc.wrtr = bufio.NewWriter(rwc)
	return sc
}

func (sc *srvIOCodec) Id() string {
	return sc.id
}

func (sc *srvIOCodec) Close() error {
	sc.wrtr.Flush()
	return sc.rwc.Close()
}

func (sc *srvIOCodec) ReadRequest() (reqId int32, funcId int16, bodySize int, err error) {
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

func (sc *srvIOCodec) ReadRequestBody(body []byte) error {
	_, err := io.ReadFull(sc.rwc, body)
	return err
}

func (sc *srvIOCodec) WriteResponse(reqId int32, opErr error, msg Encodable) error {
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
		buf := []byte(opErr.Error())
		sz := int32(len(buf))
		err = binary.Write(sc.wrtr, binary.BigEndian, sz)
		if err != nil {
			return err
		}

		err = binary.Write(sc.wrtr, binary.BigEndian, buf)
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
