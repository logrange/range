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

package sync

import (
	"context"
	"testing"
	"time"

	"github.com/logrange/range/pkg/utils/errors"
)

func BenchmarkReadLock(b *testing.B) {
	var r RWLock
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RLock()
		r.RUnlock()
	}
}

func BenchmarkLock(b *testing.B) {
	var r RWLock
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i&1 == 0 {
			r.Lock()
		} else {
			r.Unlock()
		}
	}
}

func BenchmarkTryLockUnsuccessful(b *testing.B) {
	var r RWLock
	r.Lock()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.TryRLock()
	}
}

func TestRWLockClose(t *testing.T) {
	var r RWLock
	err := r.Close()
	if err != nil {
		t.Fatal("Expecting close ok")
	}
	if r.Close() != errors.ClosedState {
		t.Fatal("Doube close must return an error")
	}
}

func TestRWLockStupid(t *testing.T) {
	var r RWLock
	r.RLock()
	go func() {
		time.Sleep(time.Millisecond)
		r.RUnlock()
	}()
	if r.wrCh == nil || r.state == stateNew {
		t.Fatal("Must be initialized ", r)
	}
	r.Lock()
	r.Unlock()
}

func TestRWLockRLock(t *testing.T) {
	var r RWLock
	r.RLock()
	if r.readers != 2 {
		t.Fatal("Expecting readers 1, but ", r)
	}
	r.RUnlock()
	r.RLock()
	r.RLock()
	if r.readers != 3 {
		t.Fatal("Expecting readers 2, but ", r)
	}
	r.RUnlock()
	r.RUnlock()
	if r.readers != 1 {
		t.Fatal("Expecting readers 0, but ", r)
	}

	r.RLock()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.RUnlock()
	}()
	r.Lock()
	r.Unlock()

	r.Lock()
	func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
			t.Fatal("Must panicing")
		}()
		r.RUnlock()
	}()
	r.Unlock()

	r.Close()
	err := r.RLock()
	if err != errors.ClosedState {
		t.Fatal("must be wrong state to acquire reader after closing err=", err)
	}
}

func TestRWLockLock(t *testing.T) {
	var r RWLock
	r.Lock()
	if r.writers != 1 || r.readers != -rwLockMaxReaders+1 {
		t.Fatal("Wrong rc=", r, ", after 1 writer acquisition")
	}

	r.Unlock()
	if r.writers != 0 || r.readers != 1 {
		t.Fatal("Wrong rc=", r, ", after writer release")
	}
	r.Lock()

	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.Unlock()
	}()
	r.Lock()
	if time.Now().Sub(start) < time.Millisecond*10 {
		t.Fatal("must be blocked here")
	}
	r.Close()
	r.Unlock()
	err := r.Lock()
	if err != errors.ClosedState {
		t.Fatal("must be wrong state to acquire writer after closing, err=", err)
	}
}

func TestRWLockLockAfterRLock(t *testing.T) {
	var r RWLock
	r.RLock()
	func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
			t.Fatal("Must panicing")
		}()
		r.Unlock()
	}()

	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.RUnlock()
	}()

	r.Lock()
	if time.Now().Sub(start) < time.Millisecond*10 {
		t.Fatal("must be blocked here")
	}
	r.Unlock()

	start = time.Now()
	r.Lock()
	go func() {
		time.Sleep(5 * time.Millisecond)
		r.Lock()
		time.Sleep(5 * time.Millisecond)
		r.Unlock()
	}()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.Unlock()
	}()

	r.Lock()
	if time.Now().Sub(start) < time.Millisecond*10 {
		t.Fatal("must be blocked here")
	}
}

func TestRWLockRLockAfterLock(t *testing.T) {
	var r RWLock

	start := time.Now()
	diff := time.Duration(0)
	r.Lock()
	go func() {
		r.Lock()
		diff = time.Now().Sub(start)
	}()
	time.Sleep(time.Millisecond * 10)
	r.Unlock()
	time.Sleep(time.Millisecond)
	if diff < time.Millisecond*10 {
		t.Fatal("Expecting diff at least 10ms, but ", diff)
	}

	// here we still have r is locked, so will try to acquire reader
	start = time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.Unlock()
	}()
	r.RLock()
	r.RLock()
	diff = time.Now().Sub(start)
	if diff < time.Millisecond*10 {
		t.Fatal("Expecting diff at least 10ms, but ", diff)
	}
	r.RUnlock()
	r.RUnlock()
}

func TestRWLockRWLockClosed(t *testing.T) {
	var r RWLock
	r.Lock()
	go func() {
		time.Sleep(time.Millisecond)
		r.Close()
	}()
	err := r.RLock()
	if err != errors.ClosedState {
		t.Fatal("Expecting err=ErrWrongState, but err=", err)
	}
	err = r.Lock()
	if err != errors.ClosedState {
		t.Fatal("Expecting err=ErrWrongState, but err=", err)
	}

	var r1 RWLock
	r1.RLock()
	go func() {
		time.Sleep(time.Millisecond)
		r1.Close()
	}()
	err = r1.Lock()
	if err != errors.ClosedState {
		t.Fatal("Expecting err=ErrWrongState, but err=", err)
	}
	err = r1.RLock()
	if err != errors.ClosedState {
		t.Fatal("Expecting err=ErrWrongState, but err=", err)
	}
}

func TestRWLockLockCtx(t *testing.T) {
	var r RWLock

	ctx, cncl := context.WithCancel(context.Background())
	r.Lock()
	if r.state != stateLocked {
		t.Fatal("Wrong state, expecting Locked, but ", r)
	}
	r.Unlock()
	if r.state != stateInit {
		t.Fatal("Wrong state, expecting Init, but ", r)
	}

	r.Lock()
	go func() {
		time.Sleep(time.Millisecond)
		r.Unlock()
	}()
	err := r.LockWithCtx(ctx)
	if err != nil || r.state != stateLocked {
		t.Fatal("Wrong state, expecting Locked, but ", r)
	}

	go func() {
		time.Sleep(time.Millisecond)
		cncl()
	}()
	err = r.LockWithCtx(ctx)
	if err != ctx.Err() || err == nil {
		t.Fatal("Expecting an err, but ", err)
	}

	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.Close()
	}()
	err = r.LockWithCtx(context.Background())
	if err != errors.ClosedState {
		t.Fatal("Must be wrong state")
	}
	if time.Now().Sub(start) < 10*time.Millisecond {
		t.Fatal("Must be at least 10 milliseconds wait")
	}
}

func TestRWLockRLockCtx(t *testing.T) {
	var r RWLock

	err := r.RLockWithCtx(nil)
	if err != nil {
		t.Fatal("Wrong state, expecting Locked, but ", r)
	}
	r.RUnlock()

	ctx, cncl := context.WithCancel(context.Background())
	r.Lock()
	go func() {
		time.Sleep(time.Millisecond)
		r.Unlock()
	}()
	err = r.RLockWithCtx(ctx)
	if err != nil || r.state != stateInit {
		t.Fatal("Wrong state, expecting Locked, but ", r)
	}
	r.RUnlock()

	r.Lock()
	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cncl()
	}()
	err = r.RLockWithCtx(ctx)
	if err != ctx.Err() || err == nil {
		t.Fatal("Expecting an err, but ", err)
	}
	if time.Now().Sub(start) < 10*time.Millisecond {
		t.Fatal("Must be at least 10 milliseconds wait")
	}

	start = time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		r.Close()
	}()
	err = r.LockWithCtx(context.Background())
	if err != errors.ClosedState {
		t.Fatal("Must be wrong state")
	}
	if time.Now().Sub(start) < 10*time.Millisecond {
		t.Fatal("Must be at least 10 milliseconds wait")
	}
}

func TestRWLockUpgradeToWrite(t *testing.T) {
	var rw RWLock
	if rw.UpgradeToWrite() {
		t.Fatal("Must not be possible to upgrade (not locked)")
	}

	rw.RLock()
	if !rw.UpgradeToWrite() {
		t.Fatal("Must be possible to upgrade")
	}
	rw.Unlock()
	rw.RLock()
	rw.RLock()
	if rw.UpgradeToWrite() {
		t.Fatal("Must not be possible to upgrade")
	}
	rw.RUnlock()
	rw.RLock()
	rw.RUnlock()
	if !rw.UpgradeToWrite() {
		t.Fatal("Must be possible to upgrade")
	}
	rw.Unlock()
}

func TestRWLockUpgradeToWrite2(t *testing.T) {
	var rw RWLock
	if rw.UpgradeToWrite() {
		t.Fatal("Must not be possible to upgrade (not locked)")
	}

	rw.RLock()
	var secondLock time.Time
	go func() {
		rw.Lock()
		secondLock = time.Now()
		rw.Unlock()
	}()

	time.Sleep(5 * time.Millisecond)
	if rw.writers != 1 || !rw.UpgradeToWrite() {
		t.Fatal("Must be possible to upgrade")
	}
	now := time.Now()
	time.Sleep(time.Millisecond)
	if now.Before(secondLock) {
		t.Fatal("oops second lock must not happen")
	}
	rw.Unlock()
	time.Sleep(time.Millisecond)
	if now.After(secondLock) {
		t.Fatal("second lock must happen after now=", now)
	}
}

func TestRWLockTryRLock(t *testing.T) {
	var rw RWLock
	if !rw.TryRLock() || rw.readers != 2 {
		t.Fatal("Must be able to acquire")
	}
	rw.RUnlock()
	rw.RLock()
	if !rw.TryRLock() || rw.readers != 3 {
		t.Fatal("Must be able to acquire")
	}
	rw.RUnlock()
	rw.RUnlock()

	rw.Lock()
	if rw.TryRLock() {
		t.Fatal("Must not be able to acquire")
	}
	rw.Unlock()

	if !rw.TryRLock() || rw.readers != 2 {
		t.Fatal("Must be able to acquire")
	}
	rw.RUnlock()
	rw.Close()
	if rw.TryRLock() {
		t.Fatal("Must not be able to acquire")
	}
}

func TestTryLock(t *testing.T) {
	var rw RWLock
	if !rw.TryLock() || rw.writers != 1 {
		t.Fatal("Must be ok")
	}
	rw.Unlock()

	rw.Lock()
	if rw.TryLock() || rw.writers != 1 {
		t.Fatal("Must not be locked")
	}
	rw.Unlock()

	rw.RLock()
	if rw.TryLock() || rw.readers != 2 {
		t.Fatal("Must not be locked")
	}
	rw.RUnlock()

	if !rw.TryLock() || rw.writers != 1 {
		t.Fatal("Must be ok")
	}
	rw.Unlock()
}

func TestDowngradeToRead(t *testing.T) {
	var rw RWLock
	rw.Lock()

	// when rw.readers == 2, it means 1 reader!!!
	if !rw.DowngradeToRead() || rw.readers != 2 || rw.writers != 0 {
		t.Fatal("Must be in read state ", rw.readers, rw.writers)
	}

	rw.RLock()
	rw.RUnlock()
	rw.RUnlock()
	rw.Lock()

	var locked int64
	go func() {
		rw.Lock()
		locked = time.Now().UnixNano()
		rw.Unlock()
	}()

	time.Sleep(time.Millisecond)
	start := time.Now().UnixNano()
	rw.DowngradeToRead()
	time.Sleep(time.Millisecond)
	rw.RUnlock()
	time.Sleep(time.Millisecond)

	if locked-start < 0 {
		t.Fatal("Lock in cycle must happen after all transformations")
	}
}
