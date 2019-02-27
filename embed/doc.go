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
Package embed provides bindings for embedding an range library in a program.

Launch an embedded range using the configuration defaults:

	package main

	import (
		"log"
		"context"
		"os"

		"github.com/logrange/range/embed"
		"github.com/logrange/range/pkg/records"
	)

	func main() {
		// Creating a new journal controller in os.Arg[1] directory
		jc, err := embed.NewJCtrlr(embed.JCtrlrConfig{JournalsDir: os.Arg[1]})
		if err != nil {
			log.Fatal("Could not create new controller, err=", err)
		}
		defer jc.Close()

		// Obtaining the journal by its name "teststream"
		j, err := jc.GetOrCreate(context.Background(), "teststream")
		if err != nil {
			log.Fatal("could not get new journal err=", err)
		}

		// Writes couple records "aaaa" and "bbbb" into the journal
		_, _, err = j.Write(context.Background(), records.SrtingsIterator("aaaa", "bbbb"))
		if err != nil {
			log.Fatal("Write err=", err)
		}

		// telling the journal to call fsync(), to be able to iterate over the fresh data right now
		j.Sync()

		// obtaining iterator to read the journal
		it := j.Iterator()
		r, _ := it.Get(context.Background())
		log.Println("First record ", string(r))

		it.Next(context.Background())
		r, _ = it.Get(context.Background())
		log.Println("Second record ", string(r))

		// jump to the first record again
		pos := it.Pos()
		pos.Idx = 0
		it.SetPos(pos)
		r, _ = it.Get(context.Background())
		log.Println("First recor ", string(r))

		// release resources
		it.Close()
	}
*/
package embed
