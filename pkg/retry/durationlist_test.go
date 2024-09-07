package retry

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func SameErrorMessages(err1, err2 error) bool {
	var msg1, msg2 string
	if err1 != nil {
		msg1 = err1.Error()
	}
	if err2 != nil {
		msg2 = err2.Error()
	}
	return msg1 == msg2
}

func TestParseSingle(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		err      error
	}{
		{
			input:    "1w2d3h4m5s",
			expected: (7*24*time.Hour + 2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second),
			err:      nil,
		},
		{
			input:    "1ms2us1ns",
			expected: (time.Millisecond + 2*time.Microsecond + time.Nanosecond),
			err:      nil,
		},
		{
			input:    "garbage",
			expected: 0,
			err:      fmt.Errorf("unexpected token \"garbage\" in duration: \"garbage\""),
		},
	}

	for _, tc := range testCases {
		result, err := ParseSingle(tc.input)
		if result != tc.expected || !SameErrorMessages(err, tc.err) {
			t.Errorf("ParseSingle(%s) = %s, %q, expected %s, %q", tc.input, result, err, tc.expected, tc.err)
		}
	}
}

func TestParseList(t *testing.T) {
	testCases := []struct {
		input    string
		expected DurationList
		err      error
	}{
		{
			input:    "1s,2m,3h",
			expected: DurationList{time.Second, 2 * time.Minute, 3 * time.Hour},
			err:      nil,
		},
		{
			input:    "500ms,1s,2s",
			expected: DurationList{500 * time.Millisecond, time.Second, 2 * time.Second},
			err:      nil,
		},
		{
			input:    "100us,200us,300us",
			expected: DurationList{100 * time.Microsecond, 200 * time.Microsecond, 300 * time.Microsecond},
			err:      nil,
		},
		{
			input:    "1h,2h,3h",
			expected: DurationList{time.Hour, 2 * time.Hour, 3 * time.Hour},
			err:      nil,
		},
		{
			input:    "1d,2d,3d",
			expected: DurationList{24 * time.Hour, 2 * 24 * time.Hour, 3 * 24 * time.Hour},
			err:      nil,
		},
		{
			input:    "1w,2w,3w",
			expected: DurationList{7 * 24 * time.Hour, 2 * 7 * 24 * time.Hour, 3 * 7 * 24 * time.Hour},
			err:      nil,
		},
		{
			input:    "garbage",
			expected: nil,
			err:      fmt.Errorf("unexpected token \"garbage\" in duration: \"garbage\""),
		},
	}

	for _, tc := range testCases {
		result, err := ParseList(tc.input)
		if !reflect.DeepEqual(result, tc.expected) || !SameErrorMessages(err, tc.err) {
			t.Errorf("ParseList(%s) = %s, %v, expected %s, %v", tc.input, result, err, tc.expected, tc.err)
		}
	}
}
