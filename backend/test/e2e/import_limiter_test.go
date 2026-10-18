package e2e

import (
	"context"
	"testing"
	"time"

	redisCache "github.com/holdennekt/sgame/backend/internal/infrastructure/cache/redis"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportLimiter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	lease := time.Minute

	newLimiter := func(t *testing.T, maxPerPod int) *redisCache.ImportLimiter {
		t.Helper()
		require.NoError(t, rds.FlushAll(ctx).Err())
		return redisCache.NewImportLimiter(rds, maxPerPod)
	}

	t.Run("one import per user", func(t *testing.T) {
		limiter := newLimiter(t, 2)

		release, err := limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)

		_, err = limiter.Acquire(ctx, "u1", lease)
		assert.ErrorIs(t, err, cache.ErrImportUserLimit)

		release()
		release2, err := limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)
		release2()
	})

	t.Run("pod limit across users", func(t *testing.T) {
		limiter := newLimiter(t, 2)

		release1, err := limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)
		release2, err := limiter.Acquire(ctx, "u2", lease)
		require.NoError(t, err)

		_, err = limiter.Acquire(ctx, "u3", lease)
		assert.ErrorIs(t, err, cache.ErrImportPodLimit)

		release1()
		release3, err := limiter.Acquire(ctx, "u3", lease)
		require.NoError(t, err)

		release2()
		release3()
	})

	t.Run("user rejection does not leak a pod slot", func(t *testing.T) {
		limiter := newLimiter(t, 2)

		release1, err := limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)

		_, err = limiter.Acquire(ctx, "u1", lease)
		require.ErrorIs(t, err, cache.ErrImportUserLimit)

		release2, err := limiter.Acquire(ctx, "u2", lease)
		require.NoError(t, err)

		release1()
		release2()
	})

	t.Run("user lock expires with its lease", func(t *testing.T) {
		limiter := newLimiter(t, 2)

		_, err := limiter.Acquire(ctx, "u1", 300*time.Millisecond)
		require.NoError(t, err)

		_, err = limiter.Acquire(ctx, "u1", lease)
		require.ErrorIs(t, err, cache.ErrImportUserLimit)

		time.Sleep(500 * time.Millisecond)
		_, err = limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)
	})

	t.Run("release survives a cancelled request context", func(t *testing.T) {
		limiter := newLimiter(t, 1)

		reqCtx, cancelReq := context.WithCancel(ctx)
		release, err := limiter.Acquire(reqCtx, "u1", lease)
		require.NoError(t, err)

		cancelReq()
		release()

		release2, err := limiter.Acquire(ctx, "u1", lease)
		require.NoError(t, err)
		release2()
	})
}
