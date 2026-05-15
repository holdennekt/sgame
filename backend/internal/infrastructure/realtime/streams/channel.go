package streams

import (
	"context"
	"encoding/json"
	"log"

	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const LAST_ID_POSTFIX = ":lastId"

type channel struct {
	client *redis.Client
	name   string
	stop   chan struct{}
}

func NewChannel(client *redis.Client, name string) realtime.Channel {
	return &channel{
		client: client,
		name:   name,
		stop:   make(chan struct{}),
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

func (c *channel) Recieve(ctx context.Context) <-chan message.Message {
	out := make(chan message.Message)

	lastId, err := c.client.Get(ctx, c.name+LAST_ID_POSTFIX).Result()
	if err != nil {
		if err == redis.Nil {
			lastId = "$"
		} else {
			log.Panicf("Cannot get internal room message lastId: %v", err)
		}
	}

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			default:
				streams, err := c.client.XRead(ctx, &redis.XReadArgs{
					Streams: []string{c.name, lastId},
					Count:   1,
					Block:   0,
				}).Result()
				if err != nil && err != redis.Nil {
					log.Println("XRead error:", err)
					continue
				}

				rdsMsg := streams[0].Messages[0]

				var msg message.Message
				if payload, ok := rdsMsg.Values["payload"].(string); ok {
					if err := json.Unmarshal([]byte(payload), &msg); err != nil {
						log.Println("Unmarshal error:", err)
						continue
					}
					out <- msg
					lastId = rdsMsg.ID
					if err := c.client.Set(ctx, c.name+LAST_ID_POSTFIX, lastId, 0).Err(); err != nil {
						log.Println("Failed to save lastId:", err)
					}
				}
			}
		}
	}()

	return out
}

func (c *channel) Close() error {
	close(c.stop)
	return nil
}
