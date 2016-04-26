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

//TODO add a String method? print - for missing ones

// The Parse function parse a common log format line from a the string s.
// Fields omitted with - (or additionally [-] and "-" for date and request
// respectively) in s will remain zeroed.
func Parse(s string) (*Line, error) {
	var l Line
	r := bufio.NewReader(strings.NewReader(s))

	// Host
	if h, err := r.ReadString(' '); err == io.EOF {
		l.host(h)
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.host(h[:len(h)-1])
	}

	// Ident
	if i, err := r.ReadString(' '); err == io.EOF {
		l.ident(i)
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.ident(i[:len(i)-1])
	}

	// AuthUser
	if a, err := r.ReadString(' '); err == io.EOF {
		l.authUser(a)
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		l.authUser(a[:len(a)-1])
	}

	// Date
	rune, _, err := r.ReadRune()
	if err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else if rune != '-' {
		if rune != '[' {
			return nil, err //TODO more error context
		}
		if d, err := r.ReadString(']'); err == io.EOF {
			return nil, err //TODO more context
		} else if err != nil {
			return nil, err //TODO more context
		} else {
			if err := l.date(d[:len(d) - 1]); err != nil {
				return nil, err //TODO more context
			}
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
	rune, _, err = r.ReadRune()
	if err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else if rune != '-' {
		if rune != '"' {
			return nil, err //TODO more error context
		}
		if req, err := r.ReadString('"'); err == io.EOF {
			return nil, err //TODO more context
		} else if err != nil {
			return nil, err //TODO more context
		} else {
			l.request(req[:len(req) - 1])
		}
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
		if err := l.status(s); err != nil {
			return nil, err //TODO more context
		}
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		if err := l.status(s[:len(s)-1]); err != nil {
			return nil, err //TODO more context
		}
	}

	// Bytes
	if b, err := r.ReadString(' '); err == io.EOF {
		if err := l.bytes(b); err != nil {
			return nil, err //TODO more context
		}
		return &l, nil
	} else if err != nil {
		return nil, err //TODO more error context
	} else {
		if err := l.bytes(b[:len(b)-1]); err != nil {
			return nil, err //TODO more context
		}
	}

	return &l, nil
}

func (l *Line) host(h string) {
	if h != "-" {
		l.Host = h
	}
}

func (l *Line) ident(i string) {
	if i != "-" {
		l.Ident = i
	}
}

func (l *Line) authUser(au string) {
	if au != "-" {
		l.AuthUser = au
	}
}

func (l *Line) date(d string) error {
	if d != "-" {
		if t, err := time.Parse(layout, d); err != nil {
			return err //TODO more context
		} else {
			l.Date = t
		}
	}
	return nil
}

func (l *Line) request(r string) {
	if r != "-" {
		l.Request = r
	}
}

func (l *Line) status(s string) error {
	if s != "-" {
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return err //TODO more context
		}
		l.Status = int(i)
	}
	return nil
}

func (l *Line) bytes(b string) error {
	if b != "-" {
		i, err := strconv.ParseInt(b, 10, 16)
		if err != nil {
			return err //TODO more context
		}
		l.Bytes = int(i)
	}
	return nil
}