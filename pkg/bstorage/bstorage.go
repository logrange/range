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

/*
 bstorage package provides BStorage interface for accessing to a byte-storage.
 It can be used for building data indexes on top of filesystem and store trees into
 a file via the BStorage interface.
*/
package bstorage

type (
	// BStorage interface provides an access to a byte storage
	BStorage interface {
		// Size returns the current storage size.
		Size() int64

		// Grow allows to increase the storage size.
		Grow(newSize int64) error

		// Read allows to read bytes by offset offs to the buffer buf.
		// offs must be in the range [0..Size). If buf is bigger than the data
		// is available the method will read as much as available.
		// The function returns number of bytes read, or an error, but not both.
		Read(offs int64, buf []byte) (int, error)

		// Write writes the data from buf to the storage starting by offs.
		// offs must be in the range [0..Size). If the buf is bigger than the
		// region or it is out of tge size range, the smaller amount will be written.
		// The function returns number of bytes written or an error, if any.
		Write(offs int64, buf []byte) (int, error)
	}
)
