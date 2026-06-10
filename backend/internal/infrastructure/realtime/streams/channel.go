package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const LAST_ID_POSTFIX = ":lastId"

type channel struct {
	client     *redis.Client
	name       string
	persistent bool
	cancel     context.CancelFunc
}

func NewChannel(client *redis.Client, name string, persistent bool) realtime.Channel {
	return &channel{
		client:     client,
		name:       name,
		persistent: persistent,
	}
}

func (c *channel) Send(ctx context.Context, msg message.Message) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}

	_, err = c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: c.name,
		Values: map[string]any{
			"payload": marshaled,
		},
	}).Result()
	if err != nil {
		return custerr.NewInternalErr(err)
	}

	return nil
}

func (c *channel) Receive(ctx context.Context) <-chan message.Message {
	ctx, c.cancel = context.WithCancel(ctx)
	out := make(chan message.Message)

	lastId, err := c.client.Get(ctx, c.name+LAST_ID_POSTFIX).Result()
	if err != nil {
		if err == redis.Nil {
			lastId = "$"
		} else {
			slog.Error("cannot get internal room message lastId", "err", err)
			panic(fmt.Sprintf("cannot get internal room message lastId: %v", err))
		}
	}

	go func() {
		defer close(out)
		for {
			streams, err := c.client.XRead(ctx, &redis.XReadArgs{
				Streams: []string{c.name, lastId},
				Count:   1,
				Block:   0,
			}).Result()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Error("XRead error", "err", err)
				continue
			}

			rdsMsg := streams[0].Messages[0]

			lastId = rdsMsg.ID
			if c.persistent {
				if err := c.client.Set(ctx, c.name+LAST_ID_POSTFIX, lastId, 0).Err(); err != nil {
					slog.Error("failed to save lastId", "err", err)
				}
			}

			var msg message.Message
			if payload, ok := rdsMsg.Values["payload"].(string); ok {
				if err := json.Unmarshal([]byte(payload), &msg); err != nil {
					slog.Error("unmarshal error", "err", err)
					continue
				}
				out <- msg
			}
		}
	}()

	return out
}

func (c *channel) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *channel) Delete(ctx context.Context) error {
	if err := c.client.Del(ctx, c.name, c.name+LAST_ID_POSTFIX).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
