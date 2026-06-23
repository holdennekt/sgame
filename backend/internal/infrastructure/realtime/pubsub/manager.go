package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	mu       sync.RWMutex
	pubsub   *redis.PubSub
	subs     map[string]map[uint64]chan message.Message
	refcount map[string]int
	lastID   atomic.Uint64
}

func NewManager(client *redis.Client) *Manager {
	m := &Manager{
		pubsub:   client.Subscribe(context.Background()),
		subs:     make(map[string]map[uint64]chan message.Message),
		refcount: make(map[string]int),
	}
	go m.run()
	return m
}

func (m *Manager) subscribe(ctx context.Context, channel string) <-chan message.Message {
	id := m.lastID.Add(1)
	ch := make(chan message.Message, 64)

	m.mu.Lock()
	if m.refcount[channel] == 0 {
		_ = m.pubsub.Subscribe(context.Background(), channel)
		m.subs[channel] = make(map[uint64]chan message.Message)
	}
	m.refcount[channel]++
	m.subs[channel][id] = ch
	m.mu.Unlock()

	go func() {
		<-ctx.Done()
		m.mu.Lock()
		delete(m.subs[channel], id)
		m.refcount[channel]--
		if m.refcount[channel] == 0 {
			_ = m.pubsub.Unsubscribe(context.Background(), channel)
			delete(m.subs, channel)
			delete(m.refcount, channel)
		}
		m.mu.Unlock()
	}()

	return ch
}

func (m *Manager) run() {
	for rdsMsg := range m.pubsub.Channel() {
		var msg message.Message
		if err := json.Unmarshal([]byte(rdsMsg.Payload), &msg); err != nil {
			slog.Error("pubsub manager: unmarshal error", "err", err)
			continue
		}

		m.mu.RLock()
		local := make([]chan message.Message, 0, len(m.subs[rdsMsg.Channel]))
		for _, ch := range m.subs[rdsMsg.Channel] {
			local = append(local, ch)
		}
		m.mu.RUnlock()

		for _, ch := range local {
			select {
			case ch <- msg:
			default:
				slog.Warn("pubsub manager: slow consumer, dropping message", "channel", rdsMsg.Channel)
			}
		}
	}
}
