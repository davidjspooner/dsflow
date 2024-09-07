package duration

import (
	"strings"
	"time"
)

type List []time.Duration

var unitMap map[string]time.Duration

func ParseList(s string) (List, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	})
	list := make(List, len(parts))
	for i, part := range parts {
		duration, err := time.ParseDuration(part)
		if err != nil {
			return nil, err
		}
		list[i] = duration
	}
	return list, nil
}

func (list List) String() string {
	if len(list) == 0 {
		return "<undefined>"
	}
	sb := strings.Builder{}
	for i, d := range list {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(d.String())
	}
	return sb.String()
}

func (list List) MarshalText() ([]byte, error) {
	return []byte(list.String()), nil
}

func (list *List) UnmarshalText(data []byte) error {
	l, err := ParseList(string(data))
	if err != nil {
		return err
	}
	*list = l
	return nil
}
