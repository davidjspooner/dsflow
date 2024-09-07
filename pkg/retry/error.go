package retry

import "fmt"

type Error struct {
	RecentAttempt error
	AbortReason   error
	Attempt       int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v, %v", e.AbortReason, e.RecentAttempt)
}
