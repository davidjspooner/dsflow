package retry

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/constraints"
)

func Sleep(ctx context.Context, d time.Duration) {
	tflog.Info(ctx, "waiting", map[string]any{"duration": Format(d)})
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func Min[T constraints.Ordered](a T, others ...T) T {
	least := a
	for _, v := range others {
		if v < least {
			least = v
		}
	}
	return least
}
func Max[T constraints.Ordered](a T, others ...T) T {
	least := a
	for _, v := range others {
		if v < least {
			least = v
		}
	}
	return least
}
