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
	"fmt"
	"github.com/logrange/range/pkg/kv"
	"sync/atomic"
	"time"
)

// kvLockProvider provides an implementation of LockProvider based on kv.Storage
type kvLockProvider struct {
	// Storage is an implementation of kv.Storage consistent key-value storage implementation
	Storage kv.Storage `inject:""`

	lease    kv.Lease
	path     string
	done     chan bool
	leaseTTL time.Duration
}

// kvLock provides an implementation of Locker object, which is returned by
// NewLocker() of kvLockProvider
type kvLock struct {
	dlp     *kvLockProvider
	key     kv.Key
	lockCh  chan bool
	lckCntr int32
	waiters int32
}

var _ LockProvider = (*kvLockProvider)(nil)
var dummyValue = []byte{0}

// NewKVLockProvider allows to create new instance of kvLockProvider.
// It expects instance of kv.Storage and a prefix key for the lockers key-space
func NewKVLockProvider(path string) *kvLockProvider {
	dlp := new(kvLockProvider)
	dlp.path = path
	dlp.leaseTTL = time.Second * 10
	return dlp
}

// Init is a part of linker.Initializer.
func (dlp *kvLockProvider) Init(ctx context.Context) (err error) {
	dlp.lease, err = dlp.Storage.Lessor().NewLease(ctx, dlp.leaseTTL, true)
	dlp.done = make(chan bool)
	return
}

// Shutdown frees resources borrowed by the kvLockProvider, implementation of
// linker.Shutdowner
func (dlp *kvLockProvider) Shutdown() {
	if dlp.lease != nil {
		_ = dlp.lease.Release()
	}
	close(dlp.done)
}

func (dlp *kvLockProvider) isDone() bool {
	select {
	case <-dlp.done:
		return true
	default:
	}
	return false
}

// NewLocker is part of LockProvider. It returns a kvLock object, which
// implements Locker
func (dlp *kvLockProvider) NewLocker(name string) Locker {
	if dlp.lease == nil {
		panic("kvLockProvider: NewLocker() is called for non-initialized component")
	}

	lock := new(kvLock)
	lock.dlp = dlp
	lock.lockCh = make(chan bool, 1)
	lock.lockCh <- true
	lock.key = kv.Key(dlp.path + name)

	return lock
}

func (l *kvLock) TryLock(ctx context.Context) bool {
	atomic.AddInt32(&l.waiters, 1)
	defer atomic.AddInt32(&l.waiters, -1)
	if err := l.tryLockInternal(); err != nil {
		return false
	}
	if _, err := l.dlp.Storage.Create(ctx, kv.Record{Key: l.key, Value: dummyValue, Lease: l.dlp.lease.Id()}); err == nil {
		return true
	}
	atomic.StoreInt32(&l.lckCntr, 0)
	l.lockCh <- true
	return false
}

// Lock is part of sync.Locker. It allows to acquire and hold the distributed lock
func (l *kvLock) Lock() {
	err := l.lockWithCtx(context.Background())
	if err != nil {
		panic("kvLock: unhandled error situation while locking err=" + err.Error())
	}
}

// Unlock releases the distributed lock
func (l *kvLock) Unlock() {
	if !atomic.CompareAndSwapInt32(&l.lckCntr, 1, 0) {
		panic("kvLock: an attempt to unlock not-locked object " + l.String())
	}

	err := l.dlp.Storage.Delete(context.Background(), l.key)
	if err != nil && err != kv.ErrNotFound {
		// some serious situation with the storage corruption or non-availability
		fmt.Printf("kvLock.Unlock(): could not read the key %s, but will release the lock: %v\n", l.String(), err)
	}

	// unlock the lock
	l.lockCh <- true
}

// String implements fmt.Stringify
func (l *kvLock) String() string {
	return fmt.Sprintf("{lckCntr: %d, key: %s, shutdown: %t, waiters: %d}",
		atomic.LoadInt32(&l.lckCntr), l.key, l.dlp.isDone(), atomic.LoadInt32(&l.waiters))
}

// lockWithCtx is part of Locker.
func (l *kvLock) lockWithCtx(ctx context.Context) error {
	atomic.AddInt32(&l.waiters, 1)
	defer atomic.AddInt32(&l.waiters, -1)
	if err := l.lockInternal(ctx); err != nil {
		return err
	}

	err := ctx.Err()
	var ver kv.Version
	for err == nil {
		ver, err = l.dlp.Storage.Create(ctx, kv.Record{Key: l.key, Value: dummyValue, Lease: l.dlp.lease.Id()})
		if err == nil {
			return nil
		}

		if err == kv.ErrAlreadyExists {
			_, _ = l.dlp.Storage.WaitForVersionChange(ctx, l.key, ver)
			err = ctx.Err()
		}
	}

	atomic.StoreInt32(&l.lckCntr, 0)
	l.lockCh <- true
	return err
}

func (l *kvLock) isLocked() bool {
	return atomic.LoadInt32(&l.lckCntr) == 1
}

func (l *kvLock) lockInternal(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.lockCh:
		if !atomic.CompareAndSwapInt32(&l.lckCntr, 0, 1) {
			panic("kvLock.lockInternal(): internal error, invalid state " + l.String())
		}
		return nil
	case <-l.dlp.done:
		return fmt.Errorf("kvLock.lockInternal(): locking mechanism is shutdown")
	}
}

func (l *kvLock) tryLockInternal() error {
	select {
	case <-l.lockCh:
		if !atomic.CompareAndSwapInt32(&l.lckCntr, 0, 1) {
			panic("kvLock.lockInternal(): internal error, invalid state " + l.String())
		}
		return nil
	case <-l.dlp.done:
		return fmt.Errorf("kvLock.lockInternal(): locking mechanism is shutdown")
	default:
		return fmt.Errorf("could not acquire")
	}
}
