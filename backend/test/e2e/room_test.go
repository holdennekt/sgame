package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/test/e2e/testhelper"
	"github.com/stretchr/testify/require"
)

func defaultRoomOptions() map[string]any {
	return map[string]any{
		"maxPlayers":                4,
		"type":                      "public",
		"readingSymbolsPerSecond":   50,
		"questionThinkingTime":      1,
		"answerThinkingTime":        1,
		"questionThinkingTimeFinal": 2,
		"falseStartAllowed":         false,
	}
}

type roomActor struct {
	session string
	userId  string
	ws      *testhelper.WSClient
}

// setupRoom creates a room with 3 guests, joins them, connects via WebSocket, and drains
// the initial room_updated broadcast.
func setupRoom(t *testing.T) (app *testhelper.TestApp, host, p1, p2 roomActor, roomId string) {
	t.Helper()
	app = newApp(t)

	hostSession, hostId := app.Guest(t, "Host")
	p1Session, p1Id := app.Guest(t, "Player1")
	p2Session, p2Id := app.Guest(t, "Player2")

	packId := insertTestPack(t, containers.MongoURI)
	roomId = app.CreateRoom(t, hostSession, "Game Room", packId, defaultRoomOptions())

	app.JoinRoom(t, hostSession, roomId)
	app.JoinRoom(t, p1Session, roomId)
	app.JoinRoom(t, p2Session, roomId)

	ctx := context.Background()
	wsPath := "/api/ws/room/" + roomId

	hostWS, err := testhelper.Dial(ctx, app.Server.URL, wsPath, hostSession)
	require.NoError(t, err)
	t.Cleanup(func() { _ = hostWS.Close() })

	p1WS, err := testhelper.Dial(ctx, app.Server.URL, wsPath, p1Session)
	require.NoError(t, err)
	t.Cleanup(func() { _ = p1WS.Close() })

	p2WS, err := testhelper.Dial(ctx, app.Server.URL, wsPath, p2Session)
	require.NoError(t, err)
	t.Cleanup(func() { _ = p2WS.Close() })

	// drain initial room_updated on each connection
	hostWS.Expect(t, domain.RoomUpdated)
	p1WS.Expect(t, domain.RoomUpdated)
	p2WS.Expect(t, domain.RoomUpdated)

	// give processor goroutines time to subscribe to the broadcast channel
	time.Sleep(50 * time.Millisecond)

	host = roomActor{hostSession, hostId, hostWS}
	p1 = roomActor{p1Session, p1Id, p1WS}
	p2 = roomActor{p2Session, p2Id, p2WS}
	return
}

// broadcastExpect asserts all actors receive the given event (draining unrelated messages).
func broadcastExpect(t *testing.T, event domain.Event, actors ...roomActor) {
	t.Helper()
	for _, a := range actors {
		a.ws.Expect(t, event)
	}
}

// ── TestRoomEventBroadcast ────────────────────────────────────────────────────

func TestRoomEventBroadcast(t *testing.T) {
	_, host, p1, p2, _ := setupRoom(t)
	ctx := context.Background()

	require.NoError(t, host.ws.Send(ctx, domain.StartGame, struct{}{}))
	drainUntilBroadcastState(t, domain.SelectingQuestion, host, p1, p2)
	broadcastExpect(t, domain.RoundDemo, host, p1, p2)
}

// ── TestFullGame ──────────────────────────────────────────────────────────────
//
// Drives every RoomState in the happy path + one wrong-answer detour:
//
//	WaitingForStart → SelectingQuestion
//	→ RevealingQuestion → ShowingQuestion
//	→ Answering (P1 wrong) → ShowingQuestion → Answering (P2 correct)
//	→ SelectingQuestion → Passing → Answering (P1 correct)
//	→ SelectingQuestion → Betting → Answering
//	→ SelectingFinalRoundCategory → FinalRoundBetting
//	→ ShowingFinalRoundQuestion → ValidatingFinalRoundAnswers → GameOver
func TestFullGame(t *testing.T) {
	_, host, p1, p2, _ := setupRoom(t)
	ctx := context.Background()

	all := []roomActor{host, p1, p2}

	// ── start_game → SelectingQuestion + RoundDemo ───────────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.StartGame, struct{}{}))
	drainUntilBroadcastState(t, domain.SelectingQuestion, all...)
	broadcastExpect(t, domain.RoundDemo, all...)

	// ── select Regular question → RevealingQuestion (after 5s question demo) ─
	require.NoError(t, host.ws.Send(ctx, domain.SelectQuestion, map[string]any{
		"category": "General", "index": 0,
	}))
	broadcastExpect(t, domain.QuestionDemo, all...)
	drainUntilBroadcastState(t, domain.RevealingQuestion, all...)

	// ── reveal timer fires → ShowingQuestion [auto] ───────────────────────────
	// "What is the answer?" = 20 chars @ 50 sym/s → ~400ms reveal
	drainUntilBroadcastState(t, domain.ShowingQuestion, all...)

	// ── P1 submits → Answering ────────────────────────────────────────────────
	require.NoError(t, p1.ws.Send(ctx, domain.StartAnswer, struct{}{}))
	drainUntilBroadcastState(t, domain.Answering, all...)

	// ── host marks wrong → ShowingQuestion ───────────────────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.ValidateAnswer, map[string]bool{"isCorrect": false}))
	drainUntilBroadcastState(t, domain.ShowingQuestion, all...)

	// ── P2 submits → Answering ────────────────────────────────────────────────
	require.NoError(t, p2.ws.Send(ctx, domain.StartAnswer, struct{}{}))
	drainUntilBroadcastState(t, domain.Answering, all...)

	// ── host marks correct → CorrectAnswerDemo → SelectingQuestion ───────────
	require.NoError(t, host.ws.Send(ctx, domain.ValidateAnswer, map[string]bool{"isCorrect": true}))
	broadcastExpect(t, domain.CorrectAnswerDemo, all...)
	drainUntilBroadcastState(t, domain.SelectingQuestion, all...)

	// ── select CatInBag → Passing (after question demo) ──────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.SelectQuestion, map[string]any{
		"category": "General", "index": 1,
	}))
	broadcastExpect(t, domain.QuestionDemo, all...)
	drainUntilBroadcastState(t, domain.Passing, all...)

	// ── host passes to P1 → Answering ────────────────────────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.PassQuestion, map[string]string{"passTo": p1.userId}))
	drainUntilBroadcastState(t, domain.Answering, all...)

	// ── host marks correct → CorrectAnswerDemo → SelectingQuestion ───────────
	require.NoError(t, host.ws.Send(ctx, domain.ValidateAnswer, map[string]bool{"isCorrect": true}))
	broadcastExpect(t, domain.CorrectAnswerDemo, all...)
	drainUntilBroadcastState(t, domain.SelectingQuestion, all...)

	// ── select Auction → Betting (after question demo) ────────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.SelectQuestion, map[string]any{
		"category": "General", "index": 2,
	}))
	broadcastExpect(t, domain.QuestionDemo, all...)
	drainUntilBroadcastState(t, domain.Betting, all...)

	// ── both players bet → Answering ─────────────────────────────────────────
	require.NoError(t, p1.ws.Send(ctx, domain.PlaceBet, map[string]int{"amount": 100}))
	require.NoError(t, p2.ws.Send(ctx, domain.PlaceBet, map[string]int{"amount": 100}))
	drainUntilBroadcastState(t, domain.Answering, all...)

	// ── host marks correct → CorrectAnswerDemo → SelectingFinalRoundCategory ─
	// (no more questions in round; server auto-starts final round)
	require.NoError(t, host.ws.Send(ctx, domain.ValidateAnswer, map[string]bool{"isCorrect": true}))
	broadcastExpect(t, domain.CorrectAnswerDemo, all...)
	drainUntilBroadcastState(t, domain.SelectingFinalRoundCategory, all...)

	// ── remove one of two categories → 1 left → auto FinalRoundBetting ───────
	require.NoError(t, host.ws.Send(ctx, domain.RemoveFinalRoundCategory, map[string]string{"category": "Final A"}))
	drainUntilBroadcastState(t, domain.FinalRoundBetting, all...)

	// ── both players place final round bets → ShowingFinalRoundQuestion ───────
	require.NoError(t, p1.ws.Send(ctx, domain.PlaceFinalRoundBet, map[string]int{"amount": 100}))
	require.NoError(t, p2.ws.Send(ctx, domain.PlaceFinalRoundBet, map[string]int{"amount": 100}))
	drainUntilBroadcastState(t, domain.ShowingFinalRoundQuestion, all...)

	// ── both submit final answer → timer expires → ValidatingFinalRoundAnswers ─
	require.NoError(t, p1.ws.Send(ctx, domain.SubmitFinalRoundAnswer, map[string]string{"answer": "anything"}))
	require.NoError(t, p2.ws.Send(ctx, domain.SubmitFinalRoundAnswer, map[string]string{"answer": "anything"}))
	drainUntilBroadcastState(t, domain.ValidatingFinalRoundAnswers, all...)

	// ── host validates both → GameOver ────────────────────────────────────────
	require.NoError(t, host.ws.Send(ctx, domain.ValidateFinalRoundAnswer, map[string]bool{"isCorrect": true}))
	require.NoError(t, host.ws.Send(ctx, domain.ValidateFinalRoundAnswer, map[string]bool{"isCorrect": true}))
	drainUntilBroadcastState(t, domain.GameOver, all...)
	broadcastExpect(t, domain.RoomDeleted, all...)

	t.Log("TestFullGame: all 12 RoomStates exercised")
}

// ── helpers ──────────────────────────────────────────────────────────────────

type playerRoomView struct {
	State domain.RoomState `json:"state"`
}

func roomStateFromMsg(t *testing.T, msg message.Message) domain.RoomState {
	t.Helper()
	var view playerRoomView
	require.NoError(t, json.Unmarshal(msg.Payload, &view))
	return view.State
}

// drainUntilBroadcastState drains room_updated messages on each actor until
// the target state is seen. Each Expect call has an internal timeout
// which covers demo durations and timer-driven transitions.
func drainUntilBroadcastState(t *testing.T, want domain.RoomState, actors ...roomActor) {
	t.Helper()
	for _, a := range actors {
		for {
			msg := a.ws.Expect(t, domain.RoomUpdated)
			if roomStateFromMsg(t, msg) == want {
				break
			}
		}
	}
}
