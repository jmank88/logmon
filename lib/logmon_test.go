package logmon

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	for _, file := range []string{
		"basic",
	} {
		in, err := os.Open("testdata/input/" + file + ".txt")
		if err != nil {
			t.Fatalf("failed to open input file %q: %s", file, err)
		}
		out, err := os.Open("testdata/output/basic.txt")
		if err != nil {
			t.Fatalf("failed to open output file %q: %s", file, err)
		}
		expected, err := ioutil.ReadAll(out)
		if err != nil {
			t.Fatalf("failed to read output file %q: ", file, err)
		}

		var b bytes.Buffer
		lm := NewLogmon(in, &b, 10 * time.Second, 2 * time.Minute, 10)
		lm.Monitor()

		if !bytes.Equal(b.Bytes(), expected) {
			t.Errorf("case %q; expected:\n%s\nbut got:\n%s", file, expected, b.String())
		}
	}
}
