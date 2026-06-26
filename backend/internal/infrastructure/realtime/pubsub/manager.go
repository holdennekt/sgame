package pubsub

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

type Manager struct {
	mu       sync.RWMutex
	pubsub   *redis.PubSub
	subs     map[string]map[uint64]chan []byte
	refcount map[string]int
	lastID   atomic.Uint64
}

func NewManager(client *redis.Client) *Manager {
	m := &Manager{
		pubsub:   client.Subscribe(context.Background()),
		subs:     make(map[string]map[uint64]chan []byte),
		refcount: make(map[string]int),
	}
	go m.run()
	return m
}

func (m *Manager) Subscribe(channel string) (<-chan []byte, uint64) {
	id := m.lastID.Add(1)
	ch := make(chan []byte, 64)

	m.mu.Lock()
	if m.refcount[channel] == 0 {
		_ = m.pubsub.Subscribe(context.Background(), channel)
		m.subs[channel] = make(map[uint64]chan []byte)
	}
	m.refcount[channel]++
	m.subs[channel][id] = ch
	m.mu.Unlock()

	return ch, id
}

func (m *Manager) Unsubscribe(channel string, id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subs[channel], id)
	m.refcount[channel]--
	if m.refcount[channel] == 0 {
		_ = m.pubsub.Unsubscribe(context.Background(), channel)
		delete(m.subs, channel)
		delete(m.refcount, channel)
	}
}

func (m *Manager) run() {
	for rdsMsg := range m.pubsub.Channel() {
		payload := []byte(rdsMsg.Payload)

		m.mu.RLock()
		local := make([]chan []byte, 0, len(m.subs[rdsMsg.Channel]))
		for _, ch := range m.subs[rdsMsg.Channel] {
			local = append(local, ch)
		}
		m.mu.RUnlock()

		for _, ch := range local {
			select {
			case ch <- payload:
			default:
				slog.Warn("pubsub manager: slow consumer, dropping message", "channel", rdsMsg.Channel)
			}
		}
	}
}
