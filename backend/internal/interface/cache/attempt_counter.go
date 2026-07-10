package cache

import (
	"context"
	"time"
)

type AttemptCounter interface {
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
}
