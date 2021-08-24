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
	"sync"
)

// LockProvider interface allows to create new Locker objects
type LockProvider interface {

	// NewLocker returns new Locker for the name provided
	NewLocker(name string) Locker
}

// Locker represents an object that supports sync.Locker and an
// extended functionality of locking mechanism in a distributed system
type Locker interface {
	sync.Locker

	// LockWithCtx allows to lock the lock object using the context provided.
	// The expected behavior is to have the object locked (result is nil), or
	// the result == ctx.Err(), what indicates that the context was cancelled
	// during the call. If the result is not nil and it is not ctx.Err(), this
	// indicates about some infrastructural errors, so this should be handled
	// accordingly.
	LockWithCtx(ctx context.Context) error
}
