[![Go Report Card](https://goreportcard.com/badge/logrange/range)](https://goreportcard.com/report/logrange/range) [![Build Status](https://travis-ci.com/logrange/range.svg?branch=master)](https://travis-ci.com/logrange/range)[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/logrange/range/blob/master/LICENSE)[![GoDoc](https://godoc.org/github.com/logrange/range/embed?status.png)](https://godoc.org/github.com/logrange/range/embed)
# range 
Range is persistent storage of streams of records. It is a horizontally-scalable, highly-available, blazinbly fast streams of records aggregation system. 

Range is designed to store unlimited number of sequentual records from thousands of sources. Due to its simplicity, range is very cost effective solution what makes it to be an effective tool for storing logging data, and sequential records. Range is used by [logrange](https://github.com/logrange/logrange) - distributed and extremely fast log aggregation system.

**Range's highlights:**
 - Streams of records data aggregation system
 - Available as embedded for stand-alone and distributed environments
 - Streams size-tolerant storage
 - Supports hundreds of thousands streams
 - Destigned to be an secured data storage

## Embedding Range as stand-alone stream storage 
Range could be used as a library. In this configuration the streams are persisted on local file-system. Records are available via journals - persisted streams of records. The following piece of code illustrates, how it works:

``` golang
package main

	import (
		"log"
		"context"
		"os"

		"github.com/logrange/range/embed"
		"github.com/logrange/range/pkg/records"
	)

	func main() {
		// Creating nes journal controller in dir by os.Arg[1]
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

		// Sync() forces journal to fsync() its internal buffers, to have just written
		// records be available right now
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
```

## Building range server
## Connecting to the range server

 
