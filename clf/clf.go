package clf

import (
	"time"
	"strings"
	"bufio"
	"io"
	"strconv"
)

const layout = "_2/Jan/2006:15:04:05 -0700"

type Line struct {
	Host string
	Ident string
	AuthUser string
	Date time.Time
	Request string
	Status int
	Bytes int
}

// The Parse function parse a common log format line from a the string s.
func Parse(s string) (*Line, error) {
	var l Line
	r := bufio.NewReader(strings.NewReader(s))

	// Host
	if h, err := r.ReadString(' '); err == io.EOF {
		l.Host = h
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.Host = h[:len(h)-1]
	}

	// Ident
	if i, err := r.ReadString(' '); err == io.EOF {
		l.Ident = i
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.Ident = i[:len(i)-1]
	}

	// AuthUser
	if a, err := r.ReadString(' '); err == io.EOF {
		l.AuthUser = a
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.AuthUser = a[:len(a)-1]
	}

	// Date
	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else if rune != '[' {
		return nil, err //TODO more error context
	}
	if d, err := r.ReadString(']'); err == io.EOF {
		return nil, err //TODO more context
	} else if err != nil {
		return nil, err //TODO more context
	} else {
		if t, err := time.Parse(layout, d[:len(d)-1]); err != nil {
			return nil, err //TODO more context
		} else {
			l.Date = t
		}
	}

	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more context
	} else if rune != ' ' {
		return nil, nil //TODO error
	}

	// Request
	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else if rune != '"' {
		return nil, err //TODO more error context
	}
	if req, err := r.ReadString('"'); err == io.EOF {
		return nil, err //TODO more context
	} else if err != nil {
		return nil, err //TODO more context
	} else {
		l.Request = req[:len(req)-1]
	}

	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more context
	} else if rune != ' ' {
		return nil, nil //TODO error
	}

	// Status
	if s, err := r.ReadString(' '); err == io.EOF {
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err //TODO more context
		}
		l.Status = int(i)
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		i, err := strconv.ParseInt(s[:len(s)-1], 10, 16)
		if err != nil {
			return nil, err //TODO more context
		}
		l.Status = int(i)
	}

	// Bytes
	if b, err := r.ReadString(' '); err == io.EOF {
		i, err := strconv.ParseInt(b, 10, 16)
		if err != nil {
			return nil, err //TODO more context
		}
		l.Bytes = int(i)
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		i, err := strconv.ParseInt(b[:len(b)-1], 10, 16)
		if err != nil {
			return nil, err //TODO more context
		}
		l.Bytes = int(i)
	}

	return &l, nil
}