package realtime

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/message"
)

type Channel interface {
	Send(ctx context.Context, msg message.Message) error
	Recieve(ctx context.Context) <-chan message.Message
	Close() error
}
