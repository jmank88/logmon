package logmon

import (
	"bufio"
	"fmt"
	"io"
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
func (l *logmon) Monitor() error {
	defer l.shutdown()

	go l.log()

	var eof bool
	for !eof {
		s, err := l.ReadString('\n')
		if err == io.EOF {
			eof = true
		} else if err != nil {
			return fmt.Errorf("failed to read line: ", err)
		} else {
			// Drop the '\n'
			s = s[:len(s)-1]
		}

		line, err := clf.Parse(s)
		if err != nil {
			return fmt.Errorf("failed to parse line %q: %s", s, err)
		}

		l.lines <- line
	}
	return nil
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
			case <-l.timeout:
				l.flushBucket()
			}
		}
	}
	l.flushBucket()
	l.done <- empty{}
}

func (l *logmon) handle(line *clf.Line) {
	if l.currentBucket.start == (time.Time{}) {
		l.newBucket(line.Date)
	}

	//TODO handle time jumps of >1bucket here
	if line.Date.After(l.currentBucket.start.Add(l.bucketDuration)) {
		l.flushBucket()
	}

	l.currentBucket.cnt++
	_, resource, _ := line.RequestFields()
	if resource != "" {
		sec := section(resource)
		if cnt, ok := l.summary[sec]; ok {
			l.summary[sec] = cnt + 1
		} else {
			l.summary[sec] = 1
		}
	}
}

//TODO doc
func (l *logmon) flushBucket() {
	fmt.Fprintf(l, "%s - %s\n", l.currentBucket.start.Format(clf.Layout), l.currentBucket.end.Format(clf.Layout))
	//TODO sort and limit
	fmt.Fprintf(l, "\tSection Hits: %v\n", l.summary)

	// TODO check for high/low traffic switch

	l.newBucket(l.currentBucket.end)
}

func (l *logmon) newBucket(s time.Time) {
	l.currentBucket = bucket{start: s, end: s.Add(l.bucketDuration)}
	l.summary = make(map[string]int)
	l.timeout = time.After(l.bucketDuration)
}

// The section function returns a resource URL's section.
// a section is defined as being what's before the second '/' in a URL. i.e.
// the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"
func section(resource string) string {
	// Skip over the schema
	schema := strings.Index(resource, "://")
	if schema == -1 {
		schema = 0
	} else {
		schema += 3
	}

	firstSlash := strings.Index(resource[schema:], "/")
	if firstSlash == -1 {
		// No path
		return resource
	}
	firstSlash += 1

	secondSlash := strings.Index(resource[schema+firstSlash:], "/")
	if secondSlash == -1 {
		// No sub-path
		return resource
	}
	return resource[:schema+firstSlash+secondSlash]
}

//TODO doc
type bucket struct {
	start, end time.Time
	cnt        int
}

type empty struct{}
