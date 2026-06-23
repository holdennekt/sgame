package streams

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/redis/go-redis/v9"
)

type StreamManager struct {
	mu       sync.RWMutex
	client   *redis.Client
	streams  map[string]*managedStream
	subIDGen atomic.Uint64

	cancelMu sync.Mutex
	cancel   context.CancelFunc

	wake chan struct{}
}

type managedStream struct {
	subs       map[uint64]chan message.Message
	refcount   int
	lastId     string
	persistent bool
}

func NewStreamManager(client *redis.Client) *StreamManager {
	m := &StreamManager{
		client:  client,
		streams: make(map[string]*managedStream),
		cancel:  func() {},
		wake:    make(chan struct{}, 1),
	}
	go m.run()
	return m
}

func (m *StreamManager) interrupt() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
	m.cancelMu.Lock()
	m.cancel()
	m.cancelMu.Unlock()
}

func (m *StreamManager) subscribe(channel string, persistent bool) (<-chan message.Message, uint64) {
	id := m.subIDGen.Add(1)
	ch := make(chan message.Message, 64)

	// Fetch lastId before acquiring the lock to avoid I/O while holding it.
	// Only used if this is the first subscriber for this channel.
	lastId := "$"
	if persistent {
		if stored, err := m.client.Get(context.Background(), channel+LAST_ID_POSTFIX).Result(); err == nil {
			lastId = stored
		}
	}

	m.mu.Lock()
	ms := m.streams[channel]
	if ms == nil {
		ms = &managedStream{
			subs:       make(map[uint64]chan message.Message),
			lastId:     lastId,
			persistent: persistent,
		}
		m.streams[channel] = ms
	}
	ms.subs[id] = ch
	ms.refcount++
	m.mu.Unlock()

	m.interrupt()

	return ch, id
}

func (m *StreamManager) unsubscribe(channel string, id uint64) {
	m.mu.Lock()
	ms := m.streams[channel]
	if ms != nil {
		delete(ms.subs, id)
		ms.refcount--
		if ms.refcount == 0 {
			delete(m.streams, channel)
		}
	}
	m.mu.Unlock()
	m.interrupt()
}

func (m *StreamManager) run() {
	for {
		m.mu.RLock()
		names := make([]string, 0, len(m.streams))
		ids := make([]string, 0, len(m.streams))
		for name, ms := range m.streams {
			names = append(names, name)
			ids = append(ids, ms.lastId)
		}
		m.mu.RUnlock()

		if len(names) == 0 {
			<-m.wake
			continue
		}

		xctx, cancel := context.WithCancel(context.Background())
		m.cancelMu.Lock()
		m.cancel = cancel
		m.cancelMu.Unlock()

		result, err := m.client.XRead(xctx, &redis.XReadArgs{
			Streams: append(names, ids...),
			Block:   0,
		}).Result()
		cancel()

		if err != nil {
			if xctx.Err() != nil {
				continue
			}
			slog.Error("stream manager: XRead error", "err", err)
			continue
		}

		m.fanOut(result)
	}
}

func (m *StreamManager) fanOut(results []redis.XStream) {
	type delivery struct {
		subs []chan message.Message
		msgs []message.Message
	}
	type persistEntry struct {
		channel string
		lastId  string
	}

	deliveries := make([]delivery, 0, len(results))
	var toPerist []persistEntry

	m.mu.Lock()
	for _, stream := range results {
		ms := m.streams[stream.Stream]
		if ms == nil {
			continue
		}
		var msgs []message.Message
		for _, rdsMsg := range stream.Messages {
			ms.lastId = rdsMsg.ID
			payload, ok := rdsMsg.Values["payload"].(string)
			if !ok {
				continue
			}
			var msg message.Message
			if err := json.Unmarshal([]byte(payload), &msg); err != nil {
				slog.Error("stream manager: unmarshal error", "err", err)
				continue
			}
			msgs = append(msgs, msg)
		}
		if len(msgs) == 0 {
			continue
		}
		if ms.persistent {
			toPerist = append(toPerist, persistEntry{stream.Stream, ms.lastId})
		}
		subs := make([]chan message.Message, 0, len(ms.subs))
		for _, ch := range ms.subs {
			subs = append(subs, ch)
		}
		deliveries = append(deliveries, delivery{subs, msgs})
	}
	m.mu.Unlock()

	for _, p := range toPerist {
		if err := m.client.Set(context.Background(), p.channel+LAST_ID_POSTFIX, p.lastId, 0).Err(); err != nil {
			slog.Error("stream manager: failed to persist lastId", "channel", p.channel, "err", err)
		}
	}

	for _, d := range deliveries {
		for _, msg := range d.msgs {
			for _, ch := range d.subs {
				select {
				case ch <- msg:
				default:
					slog.Warn("stream manager: slow consumer, dropping message")
				}
			}
		}
	}
}
