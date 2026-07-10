package gcpubsub

import (
	"context"

	"cloud.google.com/go/pubsub/v2"
)

type Consumer struct {
	sub *pubsub.Subscriber
}

func NewConsumer(ctx context.Context, projectID, subscriptionID string, maxOutstanding int) (*Consumer, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sub := client.Subscriber(subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = maxOutstanding
	return &Consumer{sub: sub}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, key string) error) error {
	err := c.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if msg.Attributes["eventType"] != "OBJECT_FINALIZE" {
			msg.Ack()
			return
		}
		key := msg.Attributes["objectId"]
		if key == "" {
			msg.Ack()
			return
		}
		if err := handler(ctx, key); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})
	if err != nil && ctx.Err() != nil {
		return nil
	}
	return err
}
