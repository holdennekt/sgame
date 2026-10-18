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
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
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

		ImportMaxConcurrentPerPod: 2,
	}

	application := app.InitializeApp(mdb, rds, storage.Raw{Storage: &NoopStorage{}}, storage.Compressed{Storage: &NoopStorage{}}, cfg)
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

// PostJSON POSTs a JSON body to an /api path; the caller closes the response body.
func (a *TestApp) PostJSON(t *testing.T, path string, body []byte) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, a.Server.URL+path, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

// Guest POSTs to /api/guest and returns the session cookie and user ID.
func (a *TestApp) Guest(t *testing.T, name string) (sessionCookie, userId string) {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"name": name})
	resp := a.PostJSON(t, "/api/guest", body)
	defer func() { _ = resp.Body.Close() }()
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

// Register POSTs to /api/register and returns the session cookie and user ID of the new non-guest user.
func (a *TestApp) Register(t *testing.T, login, password string) (sessionCookie, userId string) {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"login": login, "password": password})
	resp := a.PostJSON(t, "/api/register", body)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode, "register")

	var result struct {
		UserId string `json:"userId"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	for _, c := range resp.Cookies() {
		if c.Name == SessionCookieName {
			return c.Value, result.UserId
		}
	}
	t.Fatal("no sessionId cookie in /api/register response")
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

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, a.Server.URL+"/api/rooms", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionCookie})

	resp, err := a.Server.Client().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
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

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, fmt.Sprintf("%s/api/rooms/%s/join", a.Server.URL, roomId), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionCookie})

	resp, err := a.Server.Client().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode, "join room")
}
