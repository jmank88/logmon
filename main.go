package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/jmank88/logmon/clf"
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

	logmon := NewLogmon(input, bucketDuration, httDuration)
	logmon.monitor()
}

//TODO move to separate file
type logmon struct {
	*bufio.Reader
	//TODO doc these
	lines chan *clf.Line
	done  chan empty

	// Every bucket will be at least this long.
	bucketDuration time.Duration
	// High traffic will be measured enough recent buckets to cover at least this length.
	httDuration time.Duration

	// A slice of buckets used as a circular buffer
	buckets   []bucket
	bucketIdx int

	currentBucket bucket
	//TODO current bucket stats

	timeout <-chan time.Time

	highTrafficThreshold int
	highTraffic          bool
}

func NewLogmon(r io.Reader, bucketDuration, httDuration time.Duration) *logmon {
	bucketCnt := httDuration / bucketDuration

	return &logmon{
		Reader:         bufio.NewReader(r),
		lines:          make(chan *clf.Line),
		done:           make(chan empty),
		bucketDuration: bucketDuration,
		httDuration:    httDuration,
		buckets:        make([]bucket, bucketCnt),
	}
}

// The shutdown method signals completion to the logger, and blocks until it exits.
func (l *logmon) shutdown() {
	l.done <- empty{}
	close(l.lines)
}

func (l *logmon) monitor() {
	defer l.shutdown()

	go l.log()

	var eof bool
	for !eof {
		s, err := l.ReadString('\n')
		if err == io.EOF {
			eof = true
		} else if err != nil {
			log.Fatal("failed to read line: ", err)
		} else {
			// Drop the '\n'
			s = s[:len(s)-1]
		}

		line, err := clf.Parse(s)
		if err != nil {
			log.Fatalf("failed to parse line %q: %s", s, err)
		}

		l.lines <- line
	}
}

func (l *logmon) log() {
	l.timeout = time.After(l.bucketDuration)
	for {
		// Consume any available log lines.
		for {
			line, ok := <-l.lines
			if !ok {
				break
			}
			l.handle(line)
		}
		// Block until another line is available, or this bucket times out.
		select {
		case line := <-l.lines:
			l.handle(line)
		case end := <-l.timeout:
			l.closeBucket(end)
		}
	}
}

func (l *logmon) handle(line *clf.Line) {
	//TODO special case for first call?

	//TODO read TS, determine if entering new bucket, and new stretch
	//TODO add to stats
}

// -buckets >10s OK? (break it down into range of stuff, range of nothing, then start the new bucket)
// if we bump the end TS based on the line, then we can compare it against
//then end arg and emit an additional bucket of 0 activity...
//TODO but is this worth it?
func (l *logmon) closeBucket(end time.Time) {
	//TODO special case for first call? (since we may timeout right away)

	//TODO close bucket, log it
	// TODO check for high/log traffic switch

	//TODO start a new bucket

	//TODO reset timer
}

//TODO doc
type bucket struct {
	start, end time.Time
	cnt        int
}

type empty struct{}
