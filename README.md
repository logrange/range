[![Go Report Card](https://goreportcard.com/badge/logrange/range)](https://goreportcard.com/report/logrange/range) [![Build Status](https://travis-ci.com/logrange/range.svg?branch=master)](https://travis-ci.com/logrange/range)[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/logrange/range/blob/master/LICENSE)[![GoDoc](https://godoc.org/github.com/logrange/range/embed?status.png)](https://godoc.org/github.com/logrange/range/embed)
# range 
Range is persistent storage of streams of records. It is a horizontally-scalable, highly-available, blazinbly fast aggregation system of streams of records. 

Range is designed to store unlimited number of sequentual records from thousands of sources. Due to its simplicity, range is very cost effective solution what makes it to be an effective tool for storing logging data, dozens of millions records per second. Range is used by [logrange](https://github.com/logrange/logrange) - distributed and extremely fast log aggregation system.

**Range's highlights:**
 - Aggregation system of streams of records
 - Sstand-alone and Distributed configurations are available
 - Streams size doesn't impact system performance
 - Range is able to work with hundreds of thousands of streams
 - Created for and support Security of various systems
 - Optimized for storing dozens of millions records per second

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
```

## Building range server
## Connecting to the range server

 
## License

This project is licensed under the Apache Version 2.0 License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* GoLang IDE by [JetBrains](https://www.jetbrains.com/go/) is used for the code development
