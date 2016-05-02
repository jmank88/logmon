package logmon

import (
	"reflect"
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	is := newIntervals(3)
	for i := 0; i < 4; i++ {
		is.put(interval{cnt: i})
	}
	expected := intervals{
		slice: []interval{{cnt: 3}, {cnt: 1}, {cnt: 2}},
		idx:   1,
	}
	if !reflect.DeepEqual(is, &expected) {
		t.Errorf("expected %v but got %v", expected, *is)
	}
}

func TestAvgTraffic(t *testing.T) {
	ts := time.Now()
	is := intervals{
		slice: []interval{
			{cnt: 100},
			{
				start: ts,
				end:   ts.Add(10 * time.Second),
				cnt:   10,
			},
			{
				start: ts.Add(10 * time.Second),
				end:   ts.Add(20 * time.Second),
				cnt:   20,
			},
		},
	}
	got := is.avgTraffic(ts)
	expected := 15
	if got != expected {
		t.Errorf("expected %d but got %d", expected, got)
	}
}
