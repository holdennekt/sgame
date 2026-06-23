package testhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/holdennekt/sgame/backend/internal/app"
	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SessionCookieName = "sessionId"

type TestApp struct {
	Server *httptest.Server
	cancel context.CancelFunc
}

func NewTestApp(t *testing.T, mongoURI, redisAddr string) *TestApp {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)
	mdb := mongoClient.Database("sgame_test")

	host, port, _ := strings.Cut(redisAddr, ":")
	rds := redis.NewClient(&redis.Options{Addr: host + ":" + port})

	cfg := &config.Config{
		AppEnv:      "test",
		Host:        "localhost",
		Port:        "0",
		FrontendURL: "http://localhost:3000",

		TimeToBet:            2,
		TimeToPass:           2,
		QuestionDemoDuration: 1,
		IdleRoomTTL:          1,
	}

	application := app.InitializeApp(mdb, rds, &NoopStorage{}, cfg)
	handler := application.Start(ctx)

	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		cancel()
		srv.Close()
		_ = rds.Close()
		_ = mongoClient.Disconnect(context.Background())
	})

	return &TestApp{Server: srv, cancel: cancel}
}

// Guest POSTs to /api/guest and returns the session cookie and user ID.
func (a *TestApp) Guest(t *testing.T, name string) (sessionCookie, userId string) {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := a.Server.Client().Post(
		a.Server.URL+"/api/guest",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		UserId string `json:"userId"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	for _, c := range resp.Cookies() {
		if c.Name == SessionCookieName {
			return c.Value, result.UserId
		}
	}
	t.Fatal("no sessionId cookie in /api/guest response")
	return "", ""
}

// GuestSession is a convenience wrapper returning only the session cookie.
func (a *TestApp) GuestSession(t *testing.T, name string) string {
	session, _ := a.Guest(t, name)
	return session
}

// CreateRoom POSTs to /api/rooms and returns the room id.
func (a *TestApp) CreateRoom(t *testing.T, sessionCookie, name, packId string, options map[string]any) string {
	t.Helper()

	payload := map[string]any{
		"name":    name,
		"packId":  packId,
		"options": options,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, a.Server.URL+"/api/rooms", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionCookie})

	resp, err := a.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create room")

	var result struct {
		Id string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Id)
	return result.Id
}

// JoinRoom POSTs to /api/rooms/:id/join.
func (a *TestApp) JoinRoom(t *testing.T, sessionCookie, roomId string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/rooms/%s/join", a.Server.URL, roomId), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionCookie})

	resp, err := a.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "join room")
}
