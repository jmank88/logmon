package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jmank88/logmon/lib"
)

var (
	highTrafficThreshold = flag.Int("h", 1000, "High traffic threshold.")
	file                 = flag.String("f", "", "Optional input file, to be used in place of stdin.")
)

const (
	bucketDuration = 10 * time.Second
	httDuration    = 2 * time.Minute
)

func main() {
	flag.Parse()

	var input *os.File
	if *file == "" {
		input = os.Stdin
	} else {
		var err error
		input, err = os.Open(*file)
		if err != nil {
			log.Fatal("failed to open input file: ", err)
		}
	}

	lm := logmon.NewLogmon(input, os.Stdout, bucketDuration, httDuration, *highTrafficThreshold)
	lm.Monitor()
}
