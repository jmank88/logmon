package main

import (
	"flag"
	"log"
	"os"

	"fmt"
	"github.com/jmank88/logmon/lib"
)

var (
	htt = flag.Int("h", logmon.DefaultHighTrafficThreshold, fmt.Sprintf("High traffic threshold. A high traffic "+
		"alert will be triggered when the average traffic per %s over the last %s exceeds this value.",
		logmon.DefaultThresholdDuration, logmon.DefaultHighTrafficDuration))
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

	err := logmon.Monitor(input, os.Stdout, logmon.DefaultThresholdDuration, logmon.DefaultHighTrafficDuration, *htt)
	if err != nil {
		log.Println("monitoring quit unexpectedly: ", err)
	}
}
