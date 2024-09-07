package duration

import "time"

type Value struct {
	Duration time.Duration
}

func (v *Value) String() string {
	return v.Duration.String()
}

func (v *Value) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Value) UnmarshalText(data []byte) error {
	duration, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	v.Duration = duration
	return nil
}
