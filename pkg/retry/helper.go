package retry

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/davidjspooner/dsflow/pkg/duration"
)

func PtrTo[T any](v T) *T {
	return &v
}

type Helper struct {
	MaxAttempts int
	FastFail    []*regexp.Regexp
	Pause       time.Duration
	Interval    duration.List
	Timeout     time.Duration
}

func (rh *Helper) SetDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
	rh.SetDefaults()
	return context.WithDeadline(ctx, time.Now().Add(rh.Timeout))
}

func (rh *Helper) SetDefaults() {
	if rh.Interval == nil {
		rh.Interval = []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second}
	}
	if rh.MaxAttempts == 0 {
		rh.MaxAttempts = 1000 * 1000 * 1000 // 1 billion time should be enough
	}
	if rh.Timeout == 0 {
		rh.Timeout = 5 * time.Minute
	}
}

func (rh *Helper) Retry(ctx context.Context, fn func(ctx context.Context, attempt int) error) error {

	rh.SetDefaults()

	Sleep(ctx, rh.Pause)

	err := &Error{}

	for {
		if err.Attempt >= rh.MaxAttempts {
			err.AbortReason = fmt.Errorf("aborted after %d attempt(s)", err.Attempt)
			return err
		}
		err.AbortReason = ctx.Err()
		if err.AbortReason != nil {
			return err
		}
		if err.Attempt > 0 {
			interval := rh.Interval[Min(err.Attempt-1, len(rh.Interval)-1)]
			Sleep(ctx, interval)
			err.AbortReason = ctx.Err()
			if err.AbortReason != nil {
				return err
			}
		}
		err.Attempt++
		err.RecentAttempt = fn(ctx, err.Attempt)
		if err.RecentAttempt == nil {
			return nil
		}
		currentErrStr := err.RecentAttempt.Error()
		for _, hint := range rh.FastFail {
			if hint.FindStringIndex(currentErrStr) != nil {
				err.AbortReason = fmt.Errorf("fast fail")
				return err
			}
		}
	}
}
