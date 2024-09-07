package retry

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
	"time"
)

type DurationList []time.Duration

type unit struct {
	name string
	d    time.Duration
}

var units = []unit{
	{"ns", time.Nanosecond},
	{"us", time.Microsecond},
	{"ms", time.Millisecond},
	{"s", time.Second},
	{"m", time.Minute},
	{"h", time.Hour},
	{"d", 24 * time.Hour},
	{"w", 7 * 24 * time.Hour},
}

var unitMap map[string]time.Duration

func init() {
	sort.Slice(units, func(i, j int) bool {
		return units[i].d > units[j].d
	})
	unitMap = make(map[string]time.Duration)
	for _, u := range units {
		unitMap[u.name] = u.d
	}
}

func ParseSingle(ds string) (time.Duration, error) {
	ds = strings.TrimSpace(ds)
	if ds == "" {
		return 0, fmt.Errorf("<undefined>")
	}
	s := scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments,
		IsIdentRune: func(ch rune, i int) bool {
			return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
		},
	}
	s.Init(strings.NewReader(ds))

	var d time.Duration
	var f float64
	var err error
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		switch token {
		case scanner.Int, scanner.Float:
			f, err = strconv.ParseFloat(s.TokenText(), 64)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("unexpected token %q in duration: %q", s.TokenText(), ds)
		}
		token = s.Scan()
		if token == scanner.EOF {
			return 0, fmt.Errorf("missing unit in duration: %q", ds)
		}
		unit := s.TokenText()
		dPart, ok := unitMap[unit]
		if !ok {
			return 0, fmt.Errorf("invalid unit %s in duration: %q", unit, ds)
		}
		d += time.Duration(int64(f * float64(dPart)))
	}
	return d, nil
}

func ParseList(s string) (DurationList, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	})
	list := make(DurationList, len(parts))
	for i, part := range parts {
		duration, err := ParseSingle(part)
		if err != nil {
			return nil, err
		}
		list[i] = duration
	}
	return list, nil
}

func (list DurationList) String() string {
	if len(list) == 0 {
		return "<undefined>"
	}
	sb := strings.Builder{}
	for i, d := range list {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(Format(d))
	}
	return sb.String()
}

func (list DurationList) MarshalText() ([]byte, error) {
	return []byte(list.String()), nil
}

func (list *DurationList) UnmarshalText(data []byte) error {
	l, err := ParseList(string(data))
	if err != nil {
		return err
	}
	*list = l
	return nil
}

func Format(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	var sb strings.Builder
	if d < 0 {
		sb.WriteByte('-')
		d = -d
	}
	for _, u := range units {
		if d >= u.d {
			n := int64(d / u.d)
			d -= time.Duration(n * int64(u.d))
			sb.WriteString(strconv.FormatInt(int64(n), 10))
			sb.WriteString(u.name)
		}
	}
	return sb.String()
}
