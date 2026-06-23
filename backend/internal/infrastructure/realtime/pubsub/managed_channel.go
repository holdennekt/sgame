package pubsub

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

type managedServerChannelGetter struct {
	client  *redis.Client
	manager *Manager
}

func NewManagedServerChannelGetter(client *redis.Client, manager *Manager) realtime.ServerChannelGetter {
	return &managedServerChannelGetter{client, manager}
}

func (g *managedServerChannelGetter) Get(name string) realtime.Channel {
	return &managedChannel{client: g.client, manager: g.manager, name: name}
}

type managedChannel struct {
	client  *redis.Client
	manager *Manager
	name    string
	cancel  context.CancelFunc
}

func (c *managedChannel) Send(ctx context.Context, msg message.Message) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Publish(ctx, c.name, marshaled).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *managedChannel) Receive(outerCtx context.Context) <-chan message.Message {
	ctx, cancel := context.WithCancel(outerCtx)
	c.cancel = cancel

	inner := c.manager.subscribe(ctx, c.name)

	out := make(chan message.Message)
	go func() {
		defer func() {
			cancel()
			close(out)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-inner:
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func (c *managedChannel) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *managedChannel) Delete(_ context.Context) error { return nil }
