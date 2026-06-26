package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/pubsub"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	host, port, _ := strings.Cut(containers.RedisAddr, ":")
	rds := redis.NewClient(&redis.Options{Addr: host + ":" + port})
	t.Cleanup(func() { _ = rds.Close() })
	return rds
}

func testMsg(event domain.Event) message.Message {
	payload, _ := json.Marshal(map[string]string{"event": string(event)})
	return message.Message{Event: event, Payload: payload}
}

func TestPubSubFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := pubsub.NewManager(rds)
	getter := pubsub.NewManagedChannelGetter(rds, mgr)

	ch1 := getter.Get("test:fanout")
	ch2 := getter.Get("test:fanout")

	recv1 := ch1.Receive(ctx)
	recv2 := ch2.Receive(ctx)

	// give subscriptions time to register
	time.Sleep(50 * time.Millisecond)

	sent := testMsg(domain.RoomUpdated)
	require.NoError(t, ch1.Send(ctx, sent))

	got1 := waitMsg(t, recv1, 5*time.Second)
	got2 := waitMsg(t, recv2, 5*time.Second)

	assert.Equal(t, sent.Event, got1.Event)
	assert.Equal(t, sent.Event, got2.Event)
}

func TestPubSubIsolation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := pubsub.NewManager(rds)
	getter := pubsub.NewManagedChannelGetter(rds, mgr)

	chA := getter.Get("test:isolation:a")
	chB := getter.Get("test:isolation:b")

	recvB := chB.Receive(ctx)
	_ = chA.Receive(ctx)

	time.Sleep(50 * time.Millisecond)

	require.NoError(t, chA.Send(ctx, testMsg(domain.RoomUpdated)))

	select {
	case msg := <-recvB:
		t.Fatalf("channel B received unexpected message: %v", msg)
	case <-time.After(300 * time.Millisecond):
		// correct: nothing arrived on B
	}
}

func TestPubSubRefcount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := pubsub.NewManager(rds)
	getter := pubsub.NewManagedChannelGetter(rds, mgr)

	ch := getter.Get("test:refcount")

	subCtx, subCancel := context.WithCancel(ctx)
	recv := ch.Receive(subCtx)

	time.Sleep(50 * time.Millisecond)

	// cancel the subscriber context — channel should be cleaned up
	subCancel()

	// wait for goroutine cleanup
	time.Sleep(100 * time.Millisecond)

	// channel is now drained/closed; publishing should not panic
	_ = ch.Send(ctx, testMsg(domain.RoomUpdated))

	select {
	case _, ok := <-recv:
		// may be closed or empty — both are fine
		_ = ok
	default:
	}
}

func TestPubSubConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := pubsub.NewManager(rds)
	getter := pubsub.NewManagedChannelGetter(rds, mgr)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			subCtx, cancel := context.WithCancel(ctx)
			ch := getter.Get("test:concurrent")
			_ = ch.Receive(subCtx)
			time.Sleep(time.Duration(i) * 5 * time.Millisecond)
			cancel()
		}(i)
	}
	wg.Wait()
}

func waitMsg(t *testing.T, ch <-chan message.Message, timeout time.Duration) message.Message {
	t.Helper()
	select {
	case msg, ok := <-ch:
		if !ok {
			t.Fatal("channel closed before message arrived")
		}
		return msg
	case <-time.After(timeout):
		t.Fatal("timed out waiting for message")
		return message.Message{}
	}
}
