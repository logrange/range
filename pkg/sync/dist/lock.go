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

	// TryLock tries to acquire the lock and return immediately whether the
	// attempt was successful or not
	TryLock(ctx context.Context) bool
}
