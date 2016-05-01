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
	DefaultThresholdDuration    = 10 * time.Second
	DefaultHighTrafficDuration  = 2 * time.Minute
	DefaultHighTrafficThreshold = 10
)

// A logger holds configuration and summary information for logging.
type logger struct {
	io.Writer

	// Every interval will be at least this long.
	intervalDuration time.Duration
	// High traffic will be measured enough recent intervals to cover at least this length.
	httDuration time.Duration
	// A high traffic alert will be triggered when the average traffic per interval over the last httDuration exceeds this value.
	highTrafficThreshold int

	currentInterval interval
	// Hit counts for sections
	summary map[string]int
	//TODO more summary stats

	*intervals

	timeout <-chan time.Time

	highTraffic bool
}

// The Monitor function monitors lines read from r, and writes summaries to w.
// The frequency of logging is controlled by intervalDuration.
// Traffic exceeding highTrafficThreshold per intervalDuration over httDuration will trigger alerts.
func Monitor(r io.Reader, w io.Writer, intervalDuration, httDuration time.Duration, highTrafficThreshold int) error {
	// Lines are sent to the logger through this channel
	lines := make(chan *clf.Line)
	// Logger signals completion on this channel.
	done := make(chan empty)

	br := bufio.NewReader(r)

	intervalCnt := int(httDuration/intervalDuration) + 1
	l := &logger{
		Writer:               w,
		intervalDuration:     intervalDuration,
		httDuration:          httDuration,
		intervals:            newIntervals(intervalCnt),
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

// The log method logs summaries of data from lines until it is closed.
// It sends on done when complete.
func (l *logger) log(lines chan *clf.Line, done chan empty) {
	defer func() {
		// Flush and signal completion
		l.flushInterval()
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
			// Block until another line is available, or this interval times out.
			select {
			case line, ok := <-lines:
				if !ok {
					return
				}
				l.handle(line)
			case <-l.timeout:
				l.flushInterval()
			}
		}
	}
}

// The handle method processes a single line.
func (l *logger) handle(line *clf.Line) {
	if l.currentInterval.start == (time.Time{}) {
		l.newInterval(line.Date)
	}

	if line.Date.After(l.currentInterval.start.Add(l.intervalDuration)) {
		//TODO handle time jumps of >1 interval here
		l.flushInterval()
	}

	l.currentInterval.cnt++
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

// The flushInterval method logs a summary for the current interval and starts a new one.
func (l *logger) flushInterval() {
	fmt.Fprintf(l, "%s - %s\n", l.currentInterval.start.Format(clf.Layout), l.currentInterval.end.Format(clf.Layout))
	//TODO sort and limit
	fmt.Fprintf(l, "\tSection Hits: %v\n", l.summary)

	l.intervals.put(l.currentInterval)

	if avg := l.intervals.avgTraffic(l.currentInterval.end.Add(-l.httDuration)); avg > l.highTrafficThreshold {
		fmt.Fprintf(l, "High traffic generated an alert - hits = %d, triggered at %s\n", avg, l.currentInterval.end.Format(clf.Layout))
		l.highTraffic = true
	} else {
		if l.highTraffic {
			fmt.Fprintf(l, "Recovered from high traffic at %s\n", l.currentInterval.end.Format(clf.Layout))
			l.highTraffic = false
		}
	}

	l.newInterval(l.currentInterval.end)
}

// The newInterval method begins a new interval at start.
func (l *logger) newInterval(start time.Time) {
	l.currentInterval = interval{start: start, end: start.Add(l.intervalDuration)}
	l.summary = make(map[string]int)
	l.timeout = time.After(l.intervalDuration)
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

// An interval holds a hit count for a time range.
type interval struct {
	start, end time.Time
	cnt        int
}

// A slice of intervals used as a circular buffer.
type intervals struct {
	slice []interval
	idx   int
}

func newIntervals(cnt int) *intervals {
	return &intervals{slice: make([]interval, cnt)}
}

func (bs *intervals) put(b interval) {
	if bs.idx+1 > len(bs.slice) {
		bs.idx = 0
	}
	bs.slice[bs.idx] = b
}

// The avgTraffic function returns the average traffic per interval for
// all intervals which overlap or follow start.
func (bs *intervals) avgTraffic(start time.Time) int {
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
