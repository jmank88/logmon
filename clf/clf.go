package clf

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const layout = "_2/Jan/2006:15:04:05 -0700"

type Line struct {
	Host     string
	Ident    string
	AuthUser string
	Date     time.Time
	Request  string
	Status   int
	Bytes    int
}

// The String method formats a Line as a common log format string.
// Zeroed fields will be printed as '-'.
func (l *Line) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s",
		&str{l.Host},
		&str{l.Ident},
		&str{l.AuthUser},
		&dateStr{l.Date},
		&reqStr{l.Request},
		&intStr{l.Status},
		&intStr{l.Bytes})
}

type str struct {
	string
}

func (s *str) String() string {
	if s.string == "" {
		return "-"
	} else {
		return s.string
	}
}

type dateStr struct {
	time.Time
}

func (s *dateStr) String() string {
	if s.Time == (time.Time{}) {
		return "-"
	} else {
		return "[" + s.Time.Format(layout) + "]"
	}
}

type reqStr struct {
	string
}

func (s *reqStr) String() string {
	if s.string == "" {
		return "-"
	} else {
		return `"` + s.string + `"`
	}
}

type intStr struct {
	int
}

func (s *intStr) String() string {
	if s.int == 0 {
		return "-"
	} else {
		return strconv.FormatInt(int64(s.int), 10)
	}
}

// The RequestFields method returns request fields for a correctly formatted request,
// otherwise returns empty strings.
func (l *Line) RequestFields() (method, resource, protocol string) {
	splits := strings.Split(l.Request, " ")
	if len(splits) == 3 {
		method = splits[0]
		resource = splits[1]
		protocol = splits[2]
	}
	return
}

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
		return nil, &Err{
			cause: err,
			msg:   "failed to read host string",
		}
	} else {
		l.host(h[:len(h)-1])
	}

	// Ident
	if i, err := r.ReadString(' '); err == io.EOF {
		l.ident(i)
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read ident string",
		}
	} else {
		l.ident(i[:len(i)-1])
	}

	// AuthUser
	if a, err := r.ReadString(' '); err == io.EOF {
		l.authUser(a)
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read auth user string",
		}
	} else {
		l.authUser(a[:len(a)-1])
	}

	// Date
	rune, _, err := r.ReadRune()
	if err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read first rune of date",
		}
	} else if rune != '-' {
		if rune != '[' {
			return nil, fmt.Errorf("expected '-' or '[' at start of date, but got: %s", rune)
		}
		if d, err := r.ReadString(']'); err == io.EOF {
			return nil, &Err{
				cause: err,
				msg:   "unexpected end of file in middle of date string",
			}
		} else if err != nil {
			return nil, &Err{
				cause: err,
				msg:   "failed to read date string",
			}
		} else {
			if err := l.date(d[:len(d)-1]); err != nil {
				return nil, err
			}
		}
	}

	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read rune following date",
		}
	} else if rune != ' ' {
		return nil, fmt.Errorf("expected ' ' following date, but got: %s", rune)
	}

	// Request
	rune, _, err = r.ReadRune()
	if err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read first rune of request",
		}
	} else if rune != '-' {
		if rune != '"' {
			return nil, fmt.Errorf(`expected '-' or '"' at start of request, but got : %s`, rune)
		}
		if req, err := r.ReadString('"'); err == io.EOF {
			return nil, &Err{
				cause: err,
				msg:   "unexpected end of file in middle of request string",
			}
		} else if err != nil {
			return nil, &Err{
				cause: err,
				msg:   "failed to read request string",
			}
		} else {
			l.request(req[:len(req)-1])
		}
	}

	if rune, _, err := r.ReadRune(); err == io.EOF {
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read rune following request",
		}
	} else if rune != ' ' {
		return nil, fmt.Errorf("expected ' ' following request but got: %s", rune)
	}

	// Status
	if s, err := r.ReadString(' '); err == io.EOF {
		if err := l.status(s); err != nil {
			return nil, err
		}
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read status string",
		}
	} else {
		if err := l.status(s[:len(s)-1]); err != nil {
			return nil, err
		}
	}

	// Bytes
	if b, err := r.ReadString(' '); err == io.EOF {
		if err := l.bytes(b); err != nil {
			return nil, err
		}
		return &l, nil
	} else if err != nil {
		return nil, &Err{
			cause: err,
			msg:   "failed to read bytes string",
		}
	} else {
		if err := l.bytes(b[:len(b)-1]); err != nil {
			return nil, err
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
			return &Err{
				cause: err,
				msg:   "failed to parse date",
			}
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
		i, err := strconv.Atoi(s)
		if err != nil {
			return &Err{
				cause: err,
				msg:   "failed to parse status",
			}
		}
		l.Status = i
	}
	return nil
}

func (l *Line) bytes(b string) error {
	if b != "-" {
		i, err := strconv.Atoi(b)
		if err != nil {
			return &Err{
				cause: err,
				msg:   "failed to parse bytes",
			}
		}
		l.Bytes = i
	}
	return nil
}

type Err struct {
	cause error
	msg   string
}

func (e *Err) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.cause)
}
