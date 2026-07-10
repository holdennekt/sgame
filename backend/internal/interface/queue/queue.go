package queue

import "context"

type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, key string) error) error
}
