// Copyright 2018-2021 The logrange Authors
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
package dist

import (
	"context"
	"github.com/logrange/range/pkg/kv"
	"github.com/logrange/range/pkg/kv/inmem"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestKvLockProvider_Init(t *testing.T) {
	st := inmem.New()
	dlp := NewKVLockProvider("lp")
	dlp.Storage = st
	err := dlp.Init(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, dlp.lease)

	ls, err := st.Lessor().GetLease(dlp.lease.Id())
	assert.Nil(t, err)
	assert.NotNil(t, ls)
	assert.Equal(t, ls.Id(), dlp.lease.Id())
}

func TestKvLockProvider_Shutdown(t *testing.T) {
	dlp := newKVDLP()
	tearOff(dlp)

	ls, err := dlp.Storage.Lessor().GetLease(dlp.lease.Id())
	assert.NotNil(t, err)
	assert.Nil(t, ls)
}

func TestKvDistLock_Lock(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock := dlp.NewLocker("test").(*kvLock)

	for i := 0; i < 10; i++ {
		assert.False(t, lock.isLocked())
		_, err := dlp.Storage.Get(context.Background(), lock.key)
		assert.Equal(t, kv.ErrNotFound, err)

		lock.Lock()
		assert.True(t, lock.isLocked())
		_, err = dlp.Storage.Get(context.Background(), lock.key)
		assert.Nil(t, err)

		lock.Unlock()
	}
}

func TestKvDistLock_LockAfterShutdown(t *testing.T) {
	dlp := newKVDLP()
	lock := dlp.NewLocker("test")
	tearOff(dlp)
	err := lock.LockWithCtx(context.Background())
	assert.NotNil(t, err)
	assert.Panics(t, lock.Lock)
}

func TestKvDistLock_LockMultiple(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock := dlp.NewLocker("test").(*kvLock)

	lock.Lock()
	var start, end sync.WaitGroup
	start.Add(100)
	for i := 0; i < 100; i++ {
		end.Add(1)
		go func() {
			start.Done()
			lock.Lock()
			lock.Unlock()
			end.Done()
		}()
	}

	start.Wait()
	time.Sleep(10 * time.Millisecond)
	assert.True(t, lock.isLocked())
	assert.Equal(t, int32(100), lock.waiters)
	lock.Unlock()
	end.Wait()

	assert.False(t, lock.isLocked())
	assert.Equal(t, int32(0), lock.waiters)
}

func TestKvDistLock_CancelCtxInLock(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock1 := dlp.NewLocker("test").(*kvLock)

	// make record in kv
	lock1.Lock()

	start := time.Now()
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := lock1.LockWithCtx(ctx)
	assert.True(t, time.Now().Sub(start) >= 100*time.Millisecond)
	assert.Equal(t, ctx.Err(), err)
	assert.NotNil(t, err)
	assert.True(t, lock1.isLocked())

	lock1.Unlock()
}

func TestKvDistLock_CancelCtxWithRaise(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock1 := dlp.NewLocker("test").(*kvLock)
	lock2 := dlp.NewLocker("test").(*kvLock)

	lock1.Lock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := lock2.LockWithCtx(ctx)
	assert.Equal(t, ctx.Err(), err)

	assert.False(t, lock2.isLocked())
	assert.True(t, lock1.isLocked())
	lock1.Unlock()
}

func TestKvDistLock_LockExpiredCtx(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock1 := dlp.NewLocker("test").(*kvLock)
	lock2 := dlp.NewLocker("test").(*kvLock)

	// make record in kv
	lock1.Lock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := lock2.LockWithCtx(ctx)
	assert.Equal(t, ctx.Err(), err)
	assert.NotNil(t, err)
	assert.False(t, lock2.isLocked())

	lock1.Unlock()
}

func TestKvDistLock_RaiseLocking(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)
	lock1 := dlp.NewLocker("test").(*kvLock)
	lock2 := dlp.NewLocker("test").(*kvLock)

	lock1.Lock()

	var ack int32
	go func() {
		atomic.StoreInt32(&ack, 1)
		lock1.Lock()
		atomic.StoreInt32(&ack, 2)
		lock1.Unlock()
	}()
	for atomic.LoadInt32(&ack) == 0 {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(time.Millisecond)
	assert.True(t, lock1.isLocked())
	assert.Equal(t, int32(1), atomic.LoadInt32(&lock1.waiters))

	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		lock1.Unlock()
	}()
	lock2.Lock()
	assert.True(t, time.Now().Sub(start) >= 10*time.Millisecond)
	assert.True(t, lock2.isLocked())
	if lock1.isLocked() {
		assert.Equal(t, int32(1), atomic.LoadInt32(&ack))
	} else {
		assert.Equal(t, int32(2), atomic.LoadInt32(&ack))
	}
	lock2.Unlock()
	assert.False(t, lock2.isLocked())
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int32(2), atomic.LoadInt32(&ack))
	assert.False(t, lock1.isLocked())
}

func TestKvDistLock_Unlock(t *testing.T) {
	dlp := newKVDLP()
	defer tearOff(dlp)

	lock1 := dlp.NewLocker("test").(*kvLock)
	assert.Panics(t, lock1.Unlock)
}

func newKVDLP() *kvLockProvider {
	st := inmem.New()
	dlp := NewKVLockProvider("prefix")
	dlp.Storage = st
	dlp.Init(context.Background())
	return dlp
}

func tearOff(dlp *kvLockProvider) {
	dlp.Shutdown()
}
