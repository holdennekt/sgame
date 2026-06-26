package streams

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/pubsub"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const LAST_ID_POSTFIX = ":lastId"

type managedChannelGetter struct {
	client     *redis.Client
	manager    *pubsub.Manager
	persistent bool
}

// NewManagedChannelGetter returns a getter for non-persistent stream channels.
// Each subscriber always replays from "0-0" and the cursor is never written to
// Redis — suitable for broadcast channels with many concurrent subscribers.
func NewManagedChannelGetter(client *redis.Client, manager *pubsub.Manager) realtime.ChannelGetter {
	return &managedChannelGetter{client: client, manager: manager, persistent: false}
}

// NewPersistentManagedChannelGetter returns a getter for persistent stream channels.
// The subscriber resumes from the Redis-stored cursor and writes it back on close
// — suitable for single-subscriber channels (e.g. the internal events processor)
// that must survive server restarts without replaying already-processed events.
func NewPersistentManagedChannelGetter(client *redis.Client, manager *pubsub.Manager) realtime.ChannelGetter {
	return &managedChannelGetter{client: client, manager: manager, persistent: true}
}

func (g *managedChannelGetter) Get(name string) realtime.Channel {
	return &managedChannel{client: g.client, manager: g.manager, name: name, persistent: g.persistent}
}

type managedChannel struct {
	client     *redis.Client
	manager    *pubsub.Manager
	name       string
	persistent bool
	cancel     context.CancelFunc
}

type envelope struct {
	StreamMsgId string          `json:"stream_id"`
	Msg         message.Message `json:"msg"`
}

func (c *managedChannel) Send(ctx context.Context, msg message.Message) error {
	if c.persistent {
		return c.sendPersistent(ctx, msg)
	}
	return c.sendLive(ctx, msg)
}

func (c *managedChannel) sendLive(ctx context.Context, msg message.Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Publish(ctx, c.name, string(payload)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *managedChannel) sendPersistent(ctx context.Context, msg message.Message) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	streamMsgId, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: c.name,
		Values: map[string]any{"payload": string(msgBytes)},
	}).Result()
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	env, err := json.Marshal(envelope{StreamMsgId: streamMsgId, Msg: msg})
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Publish(ctx, c.name, string(env)).Err(); err != nil {
		slog.Error("stream channel: publish error", "channel", c.name, "err", err)
	}
	return nil
}

func (c *managedChannel) Receive(ctx context.Context) <-chan message.Message {
	if c.persistent {
		return c.receivePersistent(ctx)
	}
	return c.receiveLive(ctx)
}

// receiveLive subscribes to live Pub/Sub messages only — no XRange, no cursor.
func (c *managedChannel) receiveLive(ctx context.Context) <-chan message.Message {
	out := make(chan message.Message)
	ctx, c.cancel = context.WithCancel(ctx)

	inner, id := c.manager.Subscribe(c.name)

	go func() {
		defer func() {
			c.cancel()
			c.manager.Unsubscribe(c.name, id)
			close(out)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-inner:
				if !ok {
					return
				}
				var msg message.Message
				if err := json.Unmarshal(payload, &msg); err != nil {
					slog.Error("stream channel: unmarshal error", "err", err)
					continue
				}
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

// receivePersistent replays from the stored messages via XRange, then
// continues with live Pub/Sub messages. lastId is written to Redis after
// every processed message
func (c *managedChannel) receivePersistent(ctx context.Context) <-chan message.Message {
	out := make(chan message.Message, 64)
	ctx, c.cancel = context.WithCancel(ctx)

	inner, id := c.manager.Subscribe(c.name)

	cleanup := func() {
		c.cancel()
		c.manager.Unsubscribe(c.name, id)
		close(out)
	}

	lastId := "0-0"
	if stored, err := c.client.Get(ctx, c.name+LAST_ID_POSTFIX).Result(); err == nil {
		lastId = stored
	}

	entries, err := c.client.XRange(ctx, c.name, "("+lastId, "+").Result()
	if err != nil {
		if ctx.Err() != nil {
			cleanup()
			return out
		}
		slog.Error("stream channel: XRange error", "channel", c.name, "err", err)
	}

	for _, entry := range entries {
		payload, ok := entry.Values["payload"].(string)
		if !ok {
			continue
		}
		var msg message.Message
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			slog.Error("stream channel: unmarshal error", "err", err)
			continue
		}
		select {
		case out <- msg:
		case <-ctx.Done():
			cleanup()
			return out
		}
		lastId = entry.ID
		_ = c.client.Set(context.Background(), c.name+LAST_ID_POSTFIX, lastId, 0).Err()
	}

	go func() {
		defer cleanup()
		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-inner:
				if !ok {
					return
				}
				var env envelope
				if err := json.Unmarshal(payload, &env); err != nil {
					slog.Error("stream channel: unmarshal error", "err", err)
					continue
				}
				if streamIDLE(env.StreamMsgId, lastId) {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- env.Msg:
				}
				lastId = env.StreamMsgId
				_ = c.client.Set(context.Background(), c.name+LAST_ID_POSTFIX, lastId, 0).Err()
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

func (c *managedChannel) Delete(ctx context.Context) error {
	if err := c.client.Del(ctx, c.name, c.name+LAST_ID_POSTFIX).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

// streamIDLE reports whether stream ID a is less than or equal to b.
// Stream IDs have the format "milliseconds-sequence".
func streamIDLE(a, b string) bool {
	parse := func(id string) (uint64, uint64) {
		ms, seq, _ := strings.Cut(id, "-")
		t, _ := strconv.ParseUint(ms, 10, 64)
		s, _ := strconv.ParseUint(seq, 10, 64)
		return t, s
	}
	at, as := parse(a)
	bt, bs := parse(b)
	if at != bt {
		return at < bt
	}
	return as <= bs
}
