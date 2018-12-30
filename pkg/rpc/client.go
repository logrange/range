package rpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jrivets/log4g"
	lerrors "github.com/logrange/range/pkg/utils/errors"
	"github.com/pkg/errors"
	"io"
	"sync"
	"sync/atomic"
)

type (
	signalCh chan ServerResponse

	client struct {
		logger log4g.Logger

		lock       sync.Mutex
		inProgress map[int32]signalCh
		chCache    []signalCh
		closed     bool

		wLock sync.Mutex
		rwc   io.ReadWriteCloser
	}

	call struct {
		srf ServerRespFunc
	}
)

var caller int32

func newClient(rwc io.ReadWriteCloser) *client {
	c := new(client)
	c.logger = log4g.GetLogger("rpc.client")
	c.chCache = make([]signalCh, 10)
	c.inProgress = make(map[int32]signalCh)
	go c.readLoop()
	return c
}

func (c *client) Call(ctx context.Context, objId int64, funcId int32, m Message, srf ServerRespFunc) error {
	cid := atomic.AddInt32(&caller, 1)
	c.lock.Lock()
	if c.closed {
		c.lock.Unlock()
		c.logger.Warn("Call(): invoked for closed client: objId=", objId, ", funcId=", funcId)
		return lerrors.ClosedState
	}
	sc := c.arrangeSignalCh()
	c.inProgress[cid] = sc
	c.lock.Unlock()

	c.wLock.Lock()
	err := c.sendToWire(cid, objId, funcId, m)
	c.wLock.Unlock()

	if err != nil {
		return errors.Wrapf(err, "Call(): sendtToWire(%d, %d, %d, m)", cid, objId, funcId)
	}

	select {
	case sResp, ok := <-sc:
		if !ok {
			c.logger.Warn("Call(): signal channel was closed")
			return lerrors.ClosedState
		}
		if srf != nil {
			srf(sResp)
		}
	//TODO release sResp
	case <-ctx.Done():
		c.logger.Debug("Call(): ctx is closed before receiving result")
		err = ctx.Err()
	}

	c.lock.Lock()
	delete(c.inProgress, cid)
	// purge sc if needed
	select {
	case sResp, ok := <-sc:
		if ok {
			//TODO release sResp
			_ = sResp
		}
	default:

	}
	c.releaseSignalCh(sc)
	c.lock.Unlock()
	return err
}

func (c *client) Close() error {
	c.logger.Info("Close()")
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.close()
}

func (c *client) close() error {
	if c.closed {
		return lerrors.ClosedState
	}

	c.closed = true
	for _, sc := range c.inProgress {
		close(sc)
	}

	for _, sc := range c.chCache {
		close(sc)
	}
	c.inProgress = nil
	c.chCache = nil
	return c.rwc.Close()
}

func (c *client) closeByError(err error) {
	c.logger.Error("closeByError(): err=", err)
	if err == nil {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.close()
}

func (c *client) readLoop() {
	c.logger.Info("readLoop(): starting.")
	defer c.logger.Info("readLoop(): ending.")
	for {
		var cid int32
		err := binary.Read(c.rwc, binary.BigEndian, &cid)
		if err != nil {
			c.closeByError(err)
			return
		}

		var eCode int32
		err = binary.Read(c.rwc, binary.BigEndian, &eCode)
		if err != nil {
			c.closeByError(err)
			return
		}

		var sz int32
		err = binary.Read(c.rwc, binary.BigEndian, &sz)
		if err != nil {
			c.closeByError(err)
			return
		}

		var sResp ServerResponse
		if sz > 0 {
			sResp.EncodedMessage = make([]byte, sz) // TODO must be pool
			err = binary.Read(c.rwc, binary.BigEndian, sResp.EncodedMessage)
			if err != nil {
				c.closeByError(err)
				return
			}
		}

		if eCode > 0 {
			sResp.Error = fmt.Errorf(string(sResp.EncodedMessage))
			sResp.EncodedMessage = nil
		}

		c.lock.Lock()
		if c.closed {
			c.lock.Unlock()
			c.logger.Info("readLoop(): closed state detected")
			return
		}

		if sc, ok := c.inProgress[cid]; ok {
			sc <- sResp
		} else {
			//TODO release sResp
		}
		c.lock.Unlock()

	}
}

func (c *client) sendToWire(cid int32, objId int64, funcId int32, m Message) error {
	err := binary.Write(c.rwc, binary.BigEndian, cid)
	if err != nil {
		return err
	}

	err = binary.Write(c.rwc, binary.BigEndian, objId)
	if err != nil {
		return err
	}

	err = binary.Write(c.rwc, binary.BigEndian, funcId)
	if err != nil {
		return err
	}

	err = binary.Write(c.rwc, binary.BigEndian, m.Size())
	if err != nil {
		return err
	}

	return m.Marshal(c.rwc)
}

func (c *client) arrangeSignalCh() signalCh {
	var res signalCh
	idx := len(c.chCache) - 1
	if idx >= 0 {
		res = c.chCache[idx]
		c.chCache[idx] = nil
		c.chCache = c.chCache[:idx]
	} else {
		res = make(chan ServerResponse, 1)
	}
	return res
}

func (c *client) releaseSignalCh(sc signalCh) {
	if len(c.chCache) < cap(c.chCache) {
		c.chCache = append(c.chCache, sc)
	} else {
		close(sc)
	}
}
