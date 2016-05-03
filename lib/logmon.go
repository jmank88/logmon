package logmon

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jmank88/logmon/lib/internal/clf"
)

const (
	DefaultIntervalDuration     = 10 * time.Second
	DefaultHighTrafficDuration  = 2 * time.Minute
	DefaultHighTrafficThreshold = 10
)

// A logger holds configuration and summary information for logging.
type logger struct {
	io.Writer

	// The length of each interval.
	intervalDuration time.Duration
	// High traffic will be measured against enough recent intervals to cover at least this length.
	httDuration time.Duration
	// A high traffic alert will be triggered when the average traffic per interval over the last httDuration exceeds this value.
	highTrafficThreshold int

	currentInterval interval
	// Hit counts for sections
	sectionHits  stats
	methodHits   stats
	protocolHits stats
	statCodeHits stats

	*intervals

	timeout <-chan time.Time

	highTraffic bool
}

// The Monitor function monitors lines read from r, while writing summaries and alerts to w.
// The frequency of logging is controlled by intervalDuration.
// Traffic exceeding highTrafficThreshold per intervalDuration over httDuration will trigger alerts.
func Monitor(r io.Reader, w io.Writer, intervalDuration, httDuration time.Duration, highTrafficThreshold int) error {
	// Lines are sent to the logger through this channel.
	lines := make(chan *clf.Line)
	// The logger signals completion on this channel.
	done := make(chan empty)

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

	br := bufio.NewReader(r)

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

// The log method writes summaries and alerts based on data from lines until it is closed.
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
	if line.Date == (time.Time{}) {
		// Nowhere to put this.
		return
	}
	if l.currentInterval.start == (time.Time{}) {
		l.newInterval(line.Date)
	} else if line.Date.Before(l.currentInterval.start) {
		// Bad data. Can't go back in time.
		return
	}

	if line.Date.After(l.currentInterval.end) {
		l.flushInterval()
	}
	// Skip over intervals with no traffic
	for line.Date.After(l.currentInterval.end) {
		l.newInterval(l.currentInterval.end)
	}

	l.currentInterval.cnt++
	method, resource, protocol := line.RequestFields()
	if resource != "" {
		sec := section(resource)
		l.sectionHits.add(sec)
	}
	if method != "" {
		l.methodHits.add(method)
	}
	if protocol != "" {
		l.protocolHits.add(protocol)
	}
	if line.Status != 0 {
		l.statCodeHits.add(strconv.Itoa(line.Status))
	}
}

// The flushInterval method logs a summary for the current interval and starts a new one.
// It also may log high traffic alerts or recoveries based on historical hit counts.
func (l *logger) flushInterval() {
	fmt.Fprintf(l, "%s - %s\n", l.currentInterval.start.Format(clf.Layout), l.currentInterval.end.Format(clf.Layout))
	fmt.Fprintf(l, "\tTotal Hits: %d\n", l.currentInterval.cnt)
	fmt.Fprintf(l, "\tTop Sections: %v\n", l.sectionHits.top(5))
	fmt.Fprintf(l, "\tTop Methods: %v\n", l.methodHits.top(5))
	fmt.Fprintf(l, "\tTop Protocols: %v\n", l.protocolHits.top(5))
	fmt.Fprintf(l, "\tTop Status Codes: %v\n", l.statCodeHits.top(5))

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
	l.sectionHits = newStats()
	l.methodHits = newStats()
	l.protocolHits = newStats()
	l.statCodeHits = newStats()
	l.timeout = time.After(l.intervalDuration)
}

// The section function returns a resource URL's section.
// A section is defined as being what's before the second '/' in a URL. i.e.
// the section for "http://my.site.com/pages/create' is "http://my.site.com/pages".
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

type empty struct{}
