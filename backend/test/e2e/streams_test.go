package e2e

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, false)

	ch1 := getter.Get("stream:fanout")
	ch2 := getter.Get("stream:fanout")

	recv1 := ch1.Receive(ctx)
	recv2 := ch2.Receive(ctx)

	time.Sleep(50 * time.Millisecond)

	sent := testMsg(domain.RoomUpdated)
	require.NoError(t, ch1.Send(ctx, sent))

	got1 := waitMsg(t, recv1, 5*time.Second)
	got2 := waitMsg(t, recv2, 5*time.Second)

	assert.Equal(t, sent.Event, got1.Event)
	assert.Equal(t, sent.Event, got2.Event)
}

func TestStreamIsolation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, false)

	chA := getter.Get("stream:isolation:a")
	chB := getter.Get("stream:isolation:b")

	recvB := chB.Receive(ctx)
	_ = chA.Receive(ctx)

	time.Sleep(50 * time.Millisecond)

	require.NoError(t, chA.Send(ctx, testMsg(domain.Chat)))

	select {
	case msg := <-recvB:
		t.Fatalf("stream B received unexpected message: %v", msg)
	case <-time.After(300 * time.Millisecond):
	}
}

func TestPersistentStreamResume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, true)

	streamName := "stream:persistent:resume"

	// subscribe, receive first message, then close
	ch1 := getter.Get(streamName)
	recv1 := ch1.Receive(ctx)
	time.Sleep(50 * time.Millisecond)

	msg1 := testMsg(domain.RoomUpdated)
	require.NoError(t, ch1.Send(ctx, msg1))
	got := waitMsg(t, recv1, 5*time.Second)
	assert.Equal(t, msg1.Event, got.Event)

	// close first subscriber — lastId is persisted in Redis
	require.NoError(t, ch1.Close())
	time.Sleep(100 * time.Millisecond)

	// send a second message while nobody is subscribed
	ch2 := getter.Get(streamName)
	msg2 := testMsg(domain.GameEnded)
	require.NoError(t, ch2.Send(ctx, msg2))

	// new subscriber resumes from stored lastId — gets only msg2, not msg1
	ch3 := getter.Get(streamName)
	recv3 := ch3.Receive(ctx)
	time.Sleep(50 * time.Millisecond)

	got2 := waitMsg(t, recv3, 5*time.Second)
	assert.Equal(t, msg2.Event, got2.Event, "persistent subscriber should resume from last seen id")
}

func TestNonPersistentStreamNoReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, false)

	streamName := "stream:nonpersistent:noreplay"

	// write a message before any subscriber exists
	ch := getter.Get(streamName)
	require.NoError(t, ch.Send(ctx, testMsg(domain.RoomUpdated)))

	// subscriber joining after the message was written should NOT receive it
	ch2 := getter.Get(streamName)
	recv := ch2.Receive(ctx)
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-recv:
		t.Fatalf("non-persistent subscriber received old message unexpectedly: %v", msg)
	case <-time.After(400 * time.Millisecond):
		// correct: no replay
	}
}

func TestStreamCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, true)

	streamName := "stream:cleanup"
	ch := getter.Get(streamName)

	// write something so the stream key exists
	require.NoError(t, ch.Send(ctx, testMsg(domain.RoomUpdated)))
	time.Sleep(50 * time.Millisecond)

	require.NoError(t, ch.Delete(ctx))

	// both the stream key and lastId key should be gone
	exists, err := rds.Exists(ctx, streamName, streamName+streams.LAST_ID_POSTFIX).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestStreamConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := newRedisClient(t)
	mgr := streams.NewStreamManager(rds)
	getter := streams.NewManagedServerChannelGetter(rds, mgr, false)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			subCtx, cancel := context.WithCancel(ctx)
			ch := getter.Get("stream:concurrent")
			recv := ch.Receive(subCtx)
			time.Sleep(time.Duration(i) * 5 * time.Millisecond)
			cancel()
			for range recv {
			}
		}(i)
	}
	wg.Wait()
}
