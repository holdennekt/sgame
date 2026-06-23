package testhelper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/stretchr/testify/require"
)

const defaultTimeout = 15 * time.Second

type WSClient struct {
	conn *websocket.Conn
	msgs chan message.Message
	errs chan error
}

// Dial connects to path on server using the given sessionId cookie.
func Dial(ctx context.Context, serverURL, path, sessionCookie string) (*WSClient, error) {
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + path

	headers := http.Header{}
	headers.Set("Cookie", SessionCookieName+"="+sessionCookie)

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", wsURL, err)
	}

	c := &WSClient{
		conn: conn,
		msgs: make(chan message.Message, 64),
		errs: make(chan error, 1),
	}
	go c.readLoop(ctx)
	return c, nil
}

func (c *WSClient) readLoop(ctx context.Context) {
	defer close(c.msgs)
	for {
		var msg message.Message
		if err := wsjson.Read(ctx, c.conn, &msg); err != nil {
			select {
			case c.errs <- err:
			default:
			}
			return
		}
		select {
		case c.msgs <- msg:
		case <-ctx.Done():
			return
		}
	}
}

// Send sends a message to the server.
func (c *WSClient) Send(ctx context.Context, event domain.Event, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return wsjson.Write(ctx, c.conn, message.Message{Event: event, Payload: raw})
}

// Next blocks until a message arrives or timeout.
func (c *WSClient) Next(t *testing.T) message.Message {
	t.Helper()
	select {
	case msg, ok := <-c.msgs:
		if !ok {
			t.Fatal("ws connection closed before expected message arrived")
		}
		return msg
	case <-time.After(defaultTimeout):
		t.Fatalf("timed out waiting for ws message after %s", defaultTimeout)
		return message.Message{}
	}
}

// Expect blocks until a message with the given event arrives, discarding others.
func (c *WSClient) Expect(t *testing.T, event domain.Event) message.Message {
	t.Helper()
	deadline := time.After(defaultTimeout)
	for {
		select {
		case msg, ok := <-c.msgs:
			if !ok {
				t.Fatalf("ws closed before receiving %q", event)
			}
			if msg.Event == event {
				return msg
			}
		case <-deadline:
			t.Fatalf("timed out waiting for event %q", event)
			return message.Message{}
		}
	}
}

// ExpectPayload blocks until a message with event arrives and unmarshals its payload into dst.
func (c *WSClient) ExpectPayload(t *testing.T, event domain.Event, dst any) {
	t.Helper()
	msg := c.Expect(t, event)
	require.NoError(t, json.Unmarshal(msg.Payload, dst))
}

func (c *WSClient) Close() error {
	return c.conn.CloseNow()
}
