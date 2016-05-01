package logmon

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	r := strings.NewReader(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET http://my.site.com/pages/create HTTP/1.0" 200 2326`)
	expected := `10/Oct/2000:13:55:36 -0700 - 10/Oct/2000:13:55:36 -0700
	Section Hits: map[http://my.site.com/pages:1]
`
	var b bytes.Buffer
	lm := NewLogmon(r, &b, 10*time.Second, 2*time.Minute, 10)
	lm.Monitor()

	if b.String() != expected {
		t.Errorf("expected:\n%s\nbut got\n%s", expected, b.String())
	}
}
