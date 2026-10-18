package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bsm/redislock"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const (
	importUserLockPrefix = "import:user:"
	importReleaseTimeout = 5 * time.Second
)

type ImportLimiter struct {
	locker *redislock.Client
	slots  chan struct{}
}

func NewImportLimiter(client *redis.Client, maxPerPod int) *ImportLimiter {
	return &ImportLimiter{locker: redislock.New(client), slots: make(chan struct{}, maxPerPod)}
}

func (l *ImportLimiter) Acquire(ctx context.Context, userId string, lease time.Duration) (func(), error) {
	select {
	case l.slots <- struct{}{}:
	default:
		return nil, cache.ErrImportPodLimit
	}

	lock, err := l.locker.Obtain(ctx, importUserLockPrefix+userId, lease, nil)
	if err != nil {
		<-l.slots
		if errors.Is(err, redislock.ErrNotObtained) {
			return nil, cache.ErrImportUserLimit
		}
		return nil, custerr.NewInternalErr(err)
	}

	return func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), importReleaseTimeout)
		defer cancel()
		if err := lock.Release(releaseCtx); err != nil && !errors.Is(err, redislock.ErrLockNotHeld) {
			slog.Warn("failed to release import lock, it will expire", "err", err)
		}
		<-l.slots
	}, nil
}
