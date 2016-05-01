package logmon

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestMonitor(t *testing.T) {
	for _, file := range []string{
		"single-line",
		"multi-line",
		"multi-interval",
		"high-traffic",
	} {
		in, err := os.Open("testdata/input/" + file + ".txt")
		if err != nil {
			t.Fatalf("failed to open input file %q: %s", file, err)
		}
		out, err := os.Open("testdata/output/" + file + ".txt")
		if err != nil {
			t.Fatalf("failed to open output file %q: %s", file, err)
		}
		expected, err := ioutil.ReadAll(out)
		if err != nil {
			t.Fatalf("failed to read output file %q: ", file, err)
		}

		var b bytes.Buffer
		err = Monitor(in, &b, DefaultThresholdDuration, DefaultHighTrafficDuration, DefaultHighTrafficThreshold)
		if err != nil {
			t.Fatal("unexpected error calling Monitor: ", err)
		}

		if !bytes.Equal(b.Bytes(), expected) {
			t.Errorf("case %q; expected:\n%s\nbut got:\n%s", file, expected, b.String())
		}
	}
}

func TestSection(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected string
	}{
		{
			"http://my.site.com/pages/create",
			"http://my.site.com/pages",
		},
		{
			"my.site.com/pages/create",
			"my.site.com/pages",
		},
		{
			"http://my.site.com/pages/create/a/b/c",
			"http://my.site.com/pages",
		},
		{
			"/pages/create/a/b/c",
			"/pages",
		},
	} {
		got := section(testCase.input)
		if got != testCase.expected {
			t.Errorf("case %q: expected %q but got %q", testCase.input, testCase.expected, got)
		}
	}
}
