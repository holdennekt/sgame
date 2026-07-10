package redis

import (
	"context"
	"time"

	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

type AttemptCounter struct {
	client *redis.Client
}

func NewAttemptCounter(client *redis.Client) *AttemptCounter {
	return &AttemptCounter{client}
}

func (c *AttemptCounter) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, custerr.NewInternalErr(err)
	}
	if n == 1 {
		_ = c.client.Expire(ctx, key, ttl).Err()
	}
	return n, nil
}
