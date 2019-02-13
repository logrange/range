// Copyright 2018 The logrange Authors
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

package xbinary

import (
	"context"
	"io"
)

type (
	// Writable interface is implemented by objects that could be written into io.Writer. The objects that
	// support the interface must provide the size of the object in binary form and the function for
	// writing the object into a Writer.
	Writable interface {
		// WritableSize returns how many bytes the marshalled object form takes
		WritableSize() int

		// WriteTo allows to write (marshal) the object into the writer. If no
		// error happens, the function will write number of bytes returned by the WritableSize() function
		// it returns number of bytes written and an error, if any
		WriteTo(writer io.Writer) (int, error)
	}

	// WIterator interface allows to iterate over a collection of Writable objects
	WIterator interface {
		// Next moves the iterator current position to the next Writable object. Implementation
		// must define the order and which record should be the next.
		//
		// Next expects ctx to be used in case of the call is going to be blocked
		// some implementations can ignore this parameter if the the
		// operation can be perform without blocking the context.
		//
		// For some implementations, calling the function makes the result, returned
		// by previous call of Get() not relevant. It means, that NO results,
		// that were previously received by calling Get(), must be used after
		// the Next() call. In case of if the result is needed it must be copied
		// to another record before calling the Next() function.
		Next(ctx context.Context)

		// Get returns the current Writable object the iterator points to. If there is
		// no current object (all ones are iterated), or the collection is empty
		// the method will return io.EOF as the error.
		//
		// Get receives ctx to be used in case of the call is going to be blocked.
		// Some implementations can ignore this parameter if the the
		// operation can be perform without blocking the context.
		//
		// If error is nil, then the method returns current Writable object.
		Get(ctx context.Context) (Writable, error)
	}
)
