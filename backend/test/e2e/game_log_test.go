package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	redisCache "github.com/holdennekt/sgame/backend/internal/infrastructure/cache/redis"
	mongoDatabase "github.com/holdennekt/sgame/backend/internal/infrastructure/database/mongo"
	"github.com/holdennekt/sgame/backend/test/e2e/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func fetchGameEvents(t *testing.T, app *testhelper.TestApp, session, roomId string) (int, []domain.GameEvent) {
	t.Helper()
	resp := app.GetJSON(t, session, "/api/rooms/history/"+roomId+"/events")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	var body dto.GameEventsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return resp.StatusCode, body.Events
}

func assertFullGameLog(t *testing.T, app *testhelper.TestApp, roomId string, host, p1, p2 roomActor) {
	t.Helper()

	status, events := fetchGameEvents(t, app, host.session, roomId)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, events)

	for i, e := range events {
		require.Equal(t, i+1, e.Seq, "seq must be contiguous from 1")
	}
	assert.Equal(t, domain.EventPlayerJoined, events[0].Type)
	assert.Equal(t, domain.EventGameEnded, events[len(events)-1].Type)

	wantInOrder := []domain.GameEventType{
		domain.EventGameStarted, domain.EventRoundStarted,
		domain.EventQuestionSelected, domain.EventBuzzed, domain.EventAnswerValidated,
		domain.EventBuzzed, domain.EventAnswerValidated, domain.EventQuestionEnded,
		domain.EventQuestionSelected, domain.EventPassed, domain.EventAnswerValidated, domain.EventQuestionEnded,
		domain.EventQuestionSelected, domain.EventBetPlaced, domain.EventBetPlaced, domain.EventAuctionWon,
		domain.EventAnswerValidated, domain.EventQuestionEnded,
		domain.EventFinalRoundStarted, domain.EventFinalCategoryRemoved, domain.EventFinalCategoryChosen,
		domain.EventFinalBetPlaced, domain.EventFinalBetPlaced, domain.EventFinalQuestionStarted,
		domain.EventFinalAnswerSubmitted, domain.EventFinalAnswerSubmitted,
		domain.EventFinalAnswerValidated, domain.EventFinalAnswerValidated, domain.EventGameEnded,
	}
	next := 0
	for _, e := range events {
		if next < len(wantInOrder) && e.Type == wantInOrder[next] {
			next++
		}
	}
	assert.Equal(t, len(wantInOrder), next, "missing or out-of-order event after %v", wantInOrder[min(next, len(wantInOrder)-1)])

	var firstVerdict *domain.GameEvent
	for i := range events {
		if events[i].Type == domain.EventAnswerValidated {
			firstVerdict = &events[i]
			break
		}
	}
	require.NotNil(t, firstVerdict)
	assert.Equal(t, host.userId, firstVerdict.Actor)
	assert.Equal(t, p1.userId, firstVerdict.Data["player"])
	assert.Equal(t, false, firstVerdict.Data["correct"])
	assert.Equal(t, "moderator", firstVerdict.Data["source"])

	// final standings carry both players
	standings, ok := events[len(events)-1].Data["standings"].([]any)
	require.True(t, ok)
	assert.Len(t, standings, 2)
	_ = p2

	outsider := app.GuestSession(t, "Outsider")
	status, _ = fetchGameEvents(t, app, outsider, roomId)
	assert.Equal(t, http.StatusForbidden, status)

	status, _ = fetchGameEvents(t, app, host.session, uuid.NewString())
	assert.Equal(t, http.StatusNotFound, status)
}

func newGameLogRoom(roomId string) *domain.Room {
	return &domain.Room{
		Id:        roomId,
		CreatedBy: "host1",
		Moderator: &domain.Moderator{User: domain.User{Id: "host1"}, IsConnected: true},
		Players: []domain.Player{
			{User: domain.User{Id: "p1"}, Score: 100, IsConnected: true},
		},
		State: domain.SelectingQuestion,
	}
}

func TestSafeUpdate_AppendsEventsAtomicallyWithState(t *testing.T) {
	ctx := context.Background()
	cache := redisCache.NewRoomCache(newRedisClient(t))
	roomId := uuid.NewString()
	require.NoError(t, cache.Set(ctx, newGameLogRoom(roomId)))
	t.Cleanup(func() { _ = cache.Delete(ctx, roomId) })

	room, err := cache.SafeUpdate(ctx, roomId, func(r *domain.Room) error {
		return r.ChangeScore("host1", "p1", 250)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, room.EventSeq)

	events, err := cache.GetEvents(ctx, roomId)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, domain.EventScoreChanged, events[0].Type)
	assert.Equal(t, 1, events[0].Seq)
	assert.EqualValues(t, 100, events[0].Data["from"])
	assert.EqualValues(t, 250, events[0].Data["to"])
}

func TestSafeUpdate_FailedUpdateLeavesLogUntouched(t *testing.T) {
	ctx := context.Background()
	cache := redisCache.NewRoomCache(newRedisClient(t))
	roomId := uuid.NewString()
	require.NoError(t, cache.Set(ctx, newGameLogRoom(roomId)))
	t.Cleanup(func() { _ = cache.Delete(ctx, roomId) })

	_, err := cache.SafeUpdate(ctx, roomId, func(r *domain.Room) error {
		require.NoError(t, r.ChangeScore("host1", "p1", 999))
		return assert.AnError
	})
	require.ErrorIs(t, err, assert.AnError)

	events, err := cache.GetEvents(ctx, roomId)
	require.NoError(t, err)
	assert.Empty(t, events)

	room, err := cache.GetById(ctx, roomId)
	require.NoError(t, err)
	assert.Zero(t, room.EventSeq)
	assert.Equal(t, 100, room.Players[0].Score)
}

func TestFlushEventLog_IsIdempotentAndIncremental(t *testing.T) {
	ctx := context.Background()
	cache := redisCache.NewRoomCache(newRedisClient(t))
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(containers.MongoURI))
	require.NoError(t, err)
	t.Cleanup(func() { _ = mongoClient.Disconnect(context.Background()) })
	repo := mongoDatabase.NewGameEventRepository(mongoClient.Database("sgame_test"))

	roomId := uuid.NewString()
	require.NoError(t, cache.Set(ctx, newGameLogRoom(roomId)))
	t.Cleanup(func() { _ = cache.Delete(ctx, roomId) })

	update := func(score int) {
		_, err := cache.SafeUpdate(ctx, roomId, func(r *domain.Room) error {
			return r.ChangeScore("host1", "p1", score)
		})
		require.NoError(t, err)
	}
	update(1)
	update(2)

	require.NoError(t, server.FlushEventLog(ctx, cache, repo, roomId))
	require.NoError(t, server.FlushEventLog(ctx, cache, repo, roomId))
	stored, err := repo.GetByRoom(ctx, roomId)
	require.NoError(t, err)
	assert.Len(t, stored, 2)

	update(3)
	require.NoError(t, server.FlushEventLog(ctx, cache, repo, roomId))
	stored, err = repo.GetByRoom(ctx, roomId)
	require.NoError(t, err)
	require.Len(t, stored, 3)
	for i, e := range stored {
		assert.Equal(t, i+1, e.Seq)
	}
	assert.EqualValues(t, 3, stored[2].Data["to"])
}
