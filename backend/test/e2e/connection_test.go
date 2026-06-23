package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/test/e2e/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newApp(t *testing.T) *testhelper.TestApp {
	return testhelper.NewTestApp(t, containers.MongoURI, containers.RedisAddr)
}

func TestGuestAuth(t *testing.T) {
	app := newApp(t)

	body, _ := json.Marshal(map[string]string{"name": "Tester"})
	resp, err := app.Server.Client().Post(
		app.Server.URL+"/api/guest",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == testhelper.SessionCookieName {
			found = true
			assert.NotEmpty(t, c.Value)
		}
	}
	assert.True(t, found, "sessionId cookie should be set")
}

func TestWebSocketConnectLobby(t *testing.T) {
	app := newApp(t)
	session := app.GuestSession(t, "LobbyTester")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := testhelper.Dial(ctx, app.Server.URL, "/api/ws/lobby", session)
	require.NoError(t, err)
	defer client.Close()

	// connection established — no immediate messages required, just verify no error
}

func TestWebSocketUnauthorized(t *testing.T) {
	app := newApp(t)

	req, err := http.NewRequest(http.MethodGet, app.Server.URL+"/api/ws/lobby", nil)
	require.NoError(t, err)
	// deliberately no cookie

	resp, err := app.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocketConnectRoom(t *testing.T) {
	app := newApp(t)

	hostSession := app.GuestSession(t, "Host")
	packId := insertTestPack(t, containers.MongoURI)
	roomId := app.CreateRoom(t, hostSession, "Test Room", packId, map[string]any{
		"maxPlayers":              4,
		"type":                    "public",
		"readingSymbolsPerSecond": 50,
		"questionThinkingTime":    1,
		"answerThinkingTime":      1,
		"questionThinkingTimeFinal": 2,
		"falseStartAllowed":       false,
	})

	app.JoinRoom(t, hostSession, roomId)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// host connects via WS
	client, err := testhelper.Dial(ctx, app.Server.URL, "/api/ws/room/"+roomId, hostSession)
	require.NoError(t, err)
	defer client.Close()

	// first message is always room_updated with room state
	msg := client.Expect(t, "room_updated")
	assert.NotEmpty(t, msg.Payload)
}
