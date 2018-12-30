package rpc

import (
	"encoding/binary"
	"github.com/jrivets/log4g"
	"github.com/logrange/range/pkg/utils/errors"
	"io"
	"sync"
	"sync/atomic"
)

type (
	server struct {
		logger log4g.Logger
		lock   sync.Mutex

		closed bool
		funcs  atomic.Value
		conns  map[interface{}]*serverConn
	}

	serverConn struct {
		lock   sync.Mutex
		srvr   *server
		logger log4g.Logger
		rwc    io.ReadWriteCloser
	}
)

func newServer() *server {
	s := new(server)
	s.logger = log4g.GetLogger("rpc.server")
	s.funcs.Store(make(map[int32]ServerFunc))
	s.conns = make(map[interface{}]*serverConn)
	return s
}

func (s *server) Serve(connId string, rwc io.ReadWriteCloser) (*serverConn, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil, errors.ClosedState
	}

	sc := new(serverConn)
	sc.logger = log4g.GetLogger("rpc.serverConn").WithId("{" + connId + "}").(log4g.Logger)
	sc.rwc = rwc
	sc.srvr = s
	s.conns[sc] = sc
	go sc.readLoop()
	return sc, nil
}

func (s *server) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed {
		return errors.ClosedState
	}
	s.closed = true
	conns := s.conns
	go func() {
		for _, sc := range conns {
			sc.Close()
		}
	}()

	s.conns = nil
	return nil
}

func (s *server) onClose(sc *serverConn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed {
		return
	}

	delete(s.conns, sc)
}

func (s *server) RegisterFunc(funcId int32, srvFunc ServerFunc) {
	s.lock.Lock()
	defer s.lock.Unlock()

	funcs := s.funcs.Load().(map[int32]ServerFunc)
	newFncs := make(map[int32]ServerFunc)
	for k, v := range funcs {
		newFncs[k] = v
	}
	newFncs[funcId] = srvFunc

	s.funcs.Store(newFncs)
}

func (s *server) call(sReq ServerRequest, funcId int32) {
	funcs := s.funcs.Load().(map[int32]ServerFunc)
	if f, ok := funcs[funcId]; ok {
		if done := f(sReq); done {
			// TODO release sReq data from the pool
		}
	}
}

func (sc *serverConn) Close() error {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	if sc.rwc == nil {
		return errors.ClosedState
	}

	sc.rwc.Close()
	sc.rwc = nil
	sc.srvr.onClose(sc)
	return nil
}

func (sc *serverConn) closeByError(err error) {
	sc.logger.Warn("closeByError(): err=", err)
	sc.Close()
}

func (sc *serverConn) ServerResponse(clientId int32, srvResp ServerResponse) (err error) {
	err = sc.serverResponse(clientId, srvResp)
	if err != nil {
		sc.closeByError(err)
	}
	return
}

func (sc *serverConn) serverResponse(clientId int32, srvResp ServerResponse) (err error) {
	err = binary.Write(sc.rwc, binary.BigEndian, clientId)
	if err != nil {
		return
	}

	errCode := int32(0)
	buf := srvResp.EncodedMessage
	if srvResp.Error != nil {
		errCode = 1
		buf = []byte(srvResp.Error.Error())
	}

	err = binary.Write(sc.rwc, binary.BigEndian, errCode)
	if err != nil {
		return
	}

	err = binary.Write(sc.rwc, binary.BigEndian, int32(len(buf)))
	if err != nil {
		return
	}

	err = binary.Write(sc.rwc, binary.BigEndian, buf)
	if err != nil {
		return
	}
	return
}

func (sc *serverConn) readLoop() {
	sc.logger.Info("readLoop(): starting")
	defer sc.logger.Info("readLoop(): ending")
	for {
		sr, funcId, err := sc.readFromWire()
		if err != nil {
			sc.closeByError(err)
			return
		}
		sc.srvr.call(sr, funcId)
	}
}

func (sc *serverConn) readFromWire() (res ServerRequest, funcId int32, err error) {
	err = binary.Read(sc.rwc, binary.BigEndian, &res.ClientId)
	if err != nil {
		return
	}

	err = binary.Read(sc.rwc, binary.BigEndian, &res.ObjId)
	if err != nil {
		return
	}

	err = binary.Read(sc.rwc, binary.BigEndian, &funcId)
	if err != nil {
		return
	}

	var sz int32
	err = binary.Read(sc.rwc, binary.BigEndian, &sz)
	if err != nil {
		return
	}

	// TODO arrange data from pool
	buf := make([]byte, sz)
	err = binary.Read(sc.rwc, binary.BigEndian, buf)
	if err != nil {
		// TODO release buf
		return
	}
	res.Data = buf
	res.Conn = sc
	return
}
