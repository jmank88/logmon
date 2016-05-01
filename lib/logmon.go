package logmon

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jmank88/logmon/lib/internal/clf"
)

//TODO doc
type logmon struct {
	*bufio.Reader
	io.Writer

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
	// Hit counts for sections
	summary map[string]int
	//TODO more summary stats

	timeout <-chan time.Time

	highTrafficThreshold int
	highTraffic          bool
}

func NewLogmon(r io.Reader, w io.Writer, bucketDuration, httDuration time.Duration, highTrafficThreshold int) *logmon {
	bucketCnt := httDuration / bucketDuration

	return &logmon{
		Reader:               bufio.NewReader(r),
		Writer:               w,
		lines:                make(chan *clf.Line),
		done:                 make(chan empty),
		bucketDuration:       bucketDuration,
		httDuration:          httDuration,
		buckets:              make([]bucket, bucketCnt),
		highTrafficThreshold: highTrafficThreshold,
		summary:              make(map[string]int),
	}
}

// The shutdown method signals completion to the logger, and blocks until it exits.
func (l *logmon) shutdown() {
	// Signal logger to quit.
	close(l.lines)
	// Wait for logger.
	<-l.done
}

//TODO doc
func (l *logmon) Monitor() {
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
loop:
	for {
		select {
		// Consume any available log lines.
		case line, ok := <-l.lines:
			if !ok {
				break loop
			}
			l.handle(line)
		default:
			// Block until another line is available, or this bucket times out.
			select {
			case line, ok := <-l.lines:
				if !ok {
					break loop
				}
				l.handle(line)
			case end := <-l.timeout:
				l.flushBucket(end)
			}
		}
	}
	//TODO send now? or nothing
	l.flushBucket(time.Now())
	l.done <- empty{}
}

func (l *logmon) handle(line *clf.Line) {
	if l.currentBucket.start == (time.Time{}) {
		l.currentBucket.start = line.Date
		//TODO set end right away?
	}
	if line.Date != (time.Time{}) && line.Date.After(l.currentBucket.end) {
		l.currentBucket.end = line.Date
	}

	if line.Date.After(l.currentBucket.start.Add(l.bucketDuration)) {
		l.flushBucket(line.Date)
	}

	l.currentBucket.cnt++
	_, resource, _ := line.RequestFields()
	if resource != "" {
		if resURL, err := url.Parse(resource); err != nil {
			//TODO log bad url
		} else {
			sec := section(resURL)
			if cnt, ok := l.summary[sec]; ok {
				l.summary[sec] = cnt + 1
			} else {
				l.summary[sec] = 1
			}
		}
	}
}

// -buckets >10s OK? (break it down into range of stuff, range of nothing, then start the new bucket)
// if we bump the end TS based on the line, then we can compare it against
//then end arg and emit an additional bucket of 0 activity...
//TODO but is this worth it?

//TODO simpler: if currentBucket.end -> end > bucketDuration, then emit an empty bucket
func (l *logmon) flushBucket(nextStart time.Time) {
	if l.currentBucket.start == (time.Time{}) {
		//TODO flush empty bucket for nextStart-bucketDuration
		//TODO start new, return
	}
	if l.currentBucket.end == (time.Time{}) {
		if nextStart == (time.Time{}) {
			l.currentBucket.end = l.currentBucket.start.Add(l.bucketDuration)
		} else {
			l.currentBucket.end = nextStart
		}
	}

	fmt.Fprintf(l, "%s - %s\n", l.currentBucket.start.Format(clf.Layout), l.currentBucket.end.Format(clf.Layout))
	//TODO sort and limit
	fmt.Fprintf(l, "\tSection Hits: %v\n", l.summary)

	// TODO check for high/low traffic switch

	l.newBucket(nextStart)
}

func (l *logmon) newBucket(s time.Time) {
	l.currentBucket = bucket{start: s}
	l.summary = make(map[string]int)
	l.timeout = time.After(l.bucketDuration)
}

// The section function returns a resource URL's section.
// a section is defined as being what's before the second '/' in a URL. i.e.
// the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"
func section(resource *url.URL) string {
	var b bytes.Buffer
	if resource.Scheme != "" {
		fmt.Fprintf(&b, "%s://", resource.Scheme)
	}
	log.Println("path", resource.Path) //TODO

	var prefix string
	idx := strings.IndexRune(resource.Path[1:], '/')
	if idx == -1 {
		prefix = resource.Path
	} else {
		prefix = resource.Path[:idx+1]
	}
	fmt.Fprintf(&b, "%s%s", resource.Host, prefix)
	return b.String()
}

//TODO doc
type bucket struct {
	start, end time.Time
	cnt        int
}

type empty struct{}
