package rpc

import (
	"github.com/jrivets/log4g"
	"github.com/logrange/range/pkg/utils/errors"
	"sync"
	"sync/atomic"
)

type (
	server struct {
		logger log4g.Logger

		lock    sync.Mutex
		bufPool *pool
		closed  bool
		funcs   atomic.Value
		conns   map[interface{}]*serverConn
	}

	serverConn struct {
		lock   sync.Mutex
		srvr   *server
		codec  ServerCodec
		logger log4g.Logger
		closed int32
		wLock  sync.Mutex
	}
)

func NewServer() *server {
	s := new(server)
	s.logger = log4g.GetLogger("rpc.server")
	s.funcs.Store(make(map[int16]OnClientReqFunc))
	s.conns = make(map[interface{}]*serverConn)
	s.bufPool = new(pool)
	return s
}

func (s *server) Register(funcId int, cb OnClientReqFunc) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return errors.ClosedState
	}

	funcs := s.funcs.Load().(map[int16]OnClientReqFunc)
	newFncs := make(map[int16]OnClientReqFunc)
	for k, v := range funcs {
		newFncs[k] = v
	}
	newFncs[int16(funcId)] = cb

	s.funcs.Store(newFncs)
	return nil
}

func (s *server) Serve(srvCodec ServerCodec) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return errors.ClosedState
	}

	sc := new(serverConn)
	sc.logger = log4g.GetLogger("rpc.serverConn").WithId("{" + srvCodec.Id() + "}").(log4g.Logger)
	sc.codec = srvCodec
	sc.srvr = s
	s.conns[sc] = sc
	go sc.readLoop()
	return nil
}

func (s *server) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed {
		return errors.ClosedState
	}
	s.logger.Info("Close()")
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

func (s *server) call(funcId int16, reqId int32, reqBody []byte, sc *serverConn) {
	funcs := s.funcs.Load().(map[int16]OnClientReqFunc)
	if f, ok := funcs[funcId]; ok {
		f(reqId, reqBody, sc)
		return
	}
	s.logger.Warn("Got request ", reqId, " for unregistered function ", funcId, "skiping it")
	s.bufPool.release(reqBody)
}

func (sc *serverConn) Collect(buf []byte) {
	sc.srvr.bufPool.release(buf)
}

// SendResponse allows to send the response by the request id
func (sc *serverConn) SendResponse(reqId int32, opErr error, msg Encodable) {
	if atomic.LoadInt32(&sc.closed) != 0 {
		return
	}

	sc.wLock.Lock()
	err := sc.codec.WriteResponse(reqId, opErr, msg)
	sc.wLock.Unlock()

	if err != nil {
		sc.logger.Warn("Could not write response for reqId=", reqId, ", opErr=", opErr)
		sc.closeByError(err)
	}
}

func (sc *serverConn) Close() error {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	if atomic.LoadInt32(&sc.closed) != 0 {
		return errors.ClosedState
	}

	sc.logger.Info("Close()")
	sc.codec.Close()
	atomic.StoreInt32(&sc.closed, 1)
	sc.srvr.onClose(sc)
	return nil
}

func (sc *serverConn) closeByError(err error) {
	sc.logger.Warn("closeByError(): err=", err)
	sc.Close()
}

func (sc *serverConn) readLoop() {
	sc.logger.Info("readLoop(): starting")
	defer sc.logger.Info("readLoop(): ending")
	for atomic.LoadInt32(&sc.closed) == 0 {
		funcId, reqId, body, err := sc.readFromWire()
		if err != nil {
			sc.closeByError(err)
			return
		}
		sc.srvr.call(funcId, reqId, body, sc)
	}
}

func (sc *serverConn) readFromWire() (funcId int16, reqId int32, body []byte, err error) {
	var bsz int
	reqId, funcId, bsz, err = sc.codec.ReadRequest()
	if err != nil {
		return
	}

	body = sc.srvr.bufPool.arrange(bsz)
	err = sc.codec.ReadRequestBody(body)
	if err != nil {
		sc.srvr.bufPool.release(body)
		body = nil
	}
	return
}
