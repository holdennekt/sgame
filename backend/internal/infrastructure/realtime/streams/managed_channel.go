package streams

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

type managedStreamServerChannelGetter struct {
	client     *redis.Client
	manager    *StreamManager
	persistent bool
}

func NewManagedServerChannelGetter(client *redis.Client, manager *StreamManager, persistent bool) realtime.ServerChannelGetter {
	return &managedStreamServerChannelGetter{client, manager, persistent}
}

func (g *managedStreamServerChannelGetter) Get(name string) realtime.Channel {
	return &managedStreamChannel{client: g.client, manager: g.manager, name: name, persistent: g.persistent}
}

type managedStreamChannel struct {
	client     *redis.Client
	manager    *StreamManager
	name       string
	persistent bool
	cancel     context.CancelFunc
}

func (c *managedStreamChannel) Send(ctx context.Context, msg message.Message) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	_, err = c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: c.name,
		Values: map[string]any{"payload": marshaled},
	}).Result()
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *managedStreamChannel) Receive(ctx context.Context) <-chan message.Message {
	ctx, c.cancel = context.WithCancel(ctx)
	inner, id := c.manager.subscribe(c.name, c.persistent)

	out := make(chan message.Message)
	go func() {
		defer func() {
			c.manager.unsubscribe(c.name, id)
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

func (c *managedStreamChannel) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *managedStreamChannel) Delete(ctx context.Context) error {
	if err := c.client.Del(ctx, c.name, c.name+LAST_ID_POSTFIX).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
