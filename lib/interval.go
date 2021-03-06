package logmon

import "time"

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
	bs.slice[bs.idx] = b
	bs.idx++
	if bs.idx >= len(bs.slice) {
		bs.idx = 0
	}
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
