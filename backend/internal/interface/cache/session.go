package cache

import "context"

type Session interface {
	Get(ctx context.Context, key string) (string, error)
	GetKey(ctx context.Context, val string) (string, error)
	Set(ctx context.Context, key string, val string) error
}
