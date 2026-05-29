package pubsub

import (
	"context"
	"encoding/json"
	"log"

	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

type channel struct {
	client *redis.Client
	name   string
	pubSub *redis.PubSub
}

func NewChannel(client *redis.Client, name string) realtime.Channel {
	return &channel{client: client, name: name}
}

func (c *channel) Send(ctx context.Context, msg message.Message) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Publish(ctx, c.name, marshaled).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *channel) Recieve(ctx context.Context) <-chan message.Message {
	c.pubSub = c.client.Subscribe(ctx, c.name)
	messages := make(chan message.Message)

	go func() {
		defer func() {
			close(messages)
			c.pubSub.Close()
		}()
		rdsMessages := c.pubSub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case rdsMsg, ok := <-rdsMessages:
				if !ok {
					return
				}
				var msg message.Message
				if err := json.Unmarshal([]byte(rdsMsg.Payload), &msg); err != nil {
					log.Println("unmarshal error:", err)
					return
				}
				messages <- msg
			}
		}
	}()

	return messages
}

func (c *channel) Delete(_ context.Context) error { return nil }

func (c *channel) Close() error {
	if c.pubSub == nil {
		return nil
	}
	if err := c.pubSub.Close(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
