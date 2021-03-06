package clf

import (
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	date, err := time.Parse(Layout, "10/Oct/2000:13:55:36 -0700")
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		input    string
		expected Line
	}{
		{
			`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			Line{
				Host:     "127.0.0.1",
				Ident:    "user-identifier",
				AuthUser: "frank",
				Date:     date,
				Request:  "GET /apache_pb.gif HTTP/1.0",
				Status:   200,
				Bytes:    2326,
			},
		},
		{
			`- - - - - - -`,
			Line{},
		},
		{
			`- - - [-] "-" - -`,
			Line{},
		},
	} {
		l, err := Parse(testCase.input)
		if err != nil {
			t.Errorf("failed to parse %q: %s", testCase.input, err)
		} else if !reflect.DeepEqual(l, &testCase.expected) {
			t.Errorf("input: %q; expected %v but got %v", testCase.input, testCase.expected, l)
		}
	}
}

func TestString(t *testing.T) {
	date, err := time.Parse(Layout, "10/Oct/2000:13:55:36 -0700")
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		input    Line
		expected string
	}{
		{
			Line{},
			`- - - - - - -`,
		},
		{
			Line{
				Host:     "127.0.0.1",
				Ident:    "user-identifier",
				AuthUser: "frank",
				Date:     date,
				Request:  "GET /apache_pb.gif HTTP/1.0",
				Status:   200,
				Bytes:    2326,
			},
			`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
		},
	} {
		s := testCase.input.String()
		if s != testCase.expected {
			t.Errorf("input: (%v); expected %q but got %q", testCase.input, testCase.expected, s)
		}
	}
}
