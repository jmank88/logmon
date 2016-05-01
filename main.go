package main

import (
	"flag"
	"log"
	"os"

	"github.com/jmank88/logmon/lib"
)

var (
	htt  = flag.Int("h", logmon.DefaultHighTrafficThreshold, "High traffic threshold.")
	file = flag.String("f", "", "Optional input file, to be used in place of stdin.")
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

	lm := logmon.NewLogmon(input, os.Stdout, logmon.DefaultBucketDuration, logmon.DefaultHighTrafficDuration, *htt)
	if err := lm.Monitor(); err != nil {
		log.Println("monitoring quit unexpectedly: ", err)
	}
}
