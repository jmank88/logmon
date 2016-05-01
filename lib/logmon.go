package logmon

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jmank88/logmon/lib/internal/clf"
)

const (
	DefaultBucketDuration       = 10 * time.Second
	DefaultHighTrafficDuration  = 2 * time.Minute
	DefaultHighTrafficThreshold = 10
)

//TODO doc
type logger struct {
	io.Writer

	// Every bucket will be at least this long.
	bucketDuration time.Duration
	// High traffic will be measured enough recent buckets to cover at least this length.
	httDuration time.Duration
	// A high traffic alert will be triggered when the average traffic per bucket over the last httDuration exceeds this value.
	highTrafficThreshold int

	currentBucket bucket
	// Hit counts for sections
	summary map[string]int
	//TODO more summary stats

	*buckets

	timeout <-chan time.Time

	highTraffic bool
}

//TODO doc
func Monitor(r io.Reader, w io.Writer, bucketDuration, httDuration time.Duration, highTrafficThreshold int) error {
	// Lines are sent to the logger through this channel
	lines := make(chan *clf.Line)
	// Logger signals completion on this channel.
	done := make(chan empty)

	br := bufio.NewReader(r)

	bucketCnt := int(httDuration/bucketDuration) + 1
	l := &logger{
		Writer:               w,
		bucketDuration:       bucketDuration,
		httDuration:          httDuration,
		buckets:              newBuckets(bucketCnt),
		highTrafficThreshold: highTrafficThreshold,
	}

	go l.log(lines, done)

	defer func() {
		// Signal completion to the logger, and block until it exits.
		close(lines)
		<-done
	}()

	var eof bool
	for !eof {
		s, err := br.ReadString('\n')
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

		lines <- line
	}
	return nil
}

func (l *logger) log(lines chan *clf.Line, done chan empty) {
	defer func() {
		// Flush and signal completion
		l.flushBucket()
		done <- empty{}
	}()
	for {
		select {
		// Consume any available log lines.
		case line, ok := <-lines:
			if !ok {
				return
			}
			l.handle(line)
		default:
			// Block until another line is available, or this bucket times out.
			select {
			case line, ok := <-lines:
				if !ok {
					return
				}
				l.handle(line)
			case <-l.timeout:
				l.flushBucket()
			}
		}
	}
}

func (l *logger) handle(line *clf.Line) {
	if l.currentBucket.start == (time.Time{}) {
		l.newBucket(line.Date)
	}

	if line.Date.After(l.currentBucket.start.Add(l.bucketDuration)) {
		//TODO handle time jumps of >1bucket here
		//if after 2*bucketDuration
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
func (l *logger) flushBucket() {
	fmt.Fprintf(l, "%s - %s\n", l.currentBucket.start.Format(clf.Layout), l.currentBucket.end.Format(clf.Layout))
	//TODO sort and limit
	fmt.Fprintf(l, "\tSection Hits: %v\n", l.summary)

	l.buckets.put(l.currentBucket)

	if avg := l.buckets.avgTraffic(l.currentBucket.end.Add(-l.httDuration)); avg > l.highTrafficThreshold {
		fmt.Fprintf(l, "High traffic generated an alert - hits = %d, triggered at %s\n", avg, l.currentBucket.end.Format(clf.Layout))
		l.highTraffic = true
	} else {
		if l.highTraffic {
			fmt.Fprintf(l, "Recovered from high traffic at %s\n", l.currentBucket.end.Format(clf.Layout))
			l.highTraffic = false
		}
	}

	l.newBucket(l.currentBucket.end)
}

func (l *logger) newBucket(s time.Time) {
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

// A slice of buckets used as a circular buffer
type buckets struct {
	slice []bucket
	idx   int
}

func newBuckets(cnt int) *buckets {
	return &buckets{slice: make([]bucket, cnt)}
}

func (bs *buckets) put(b bucket) {
	if bs.idx+1 > len(bs.slice) {
		bs.idx = 0
	}
	bs.slice[bs.idx] = b
}

// The avgTraffic function returns the average traffic per bucket for
// all buckets which overlap or follow start.
func (bs *buckets) avgTraffic(start time.Time) int {
	var sum, cnt int
	for _, b := range bs.slice {
		if b.end.After(start) {
			sum += b.cnt
			cnt++
		}
	}
	if cnt == 0 {
		return 0
	}
	return sum / cnt
}

type empty struct{}
