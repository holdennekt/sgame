package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrImportUserLimit = errors.New("user import limit reached")
	ErrImportPodLimit  = errors.New("pod import limit reached")
)

type ImportLimiter interface {
	// Acquire allows one import per user across pods and a capped number per pod; the user lock expires after lease
	Acquire(ctx context.Context, userId string, lease time.Duration) (release func(), err error)
}
