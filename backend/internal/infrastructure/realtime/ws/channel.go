package ws

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type channel struct {
	conn *websocket.Conn
}

func NewChannel(conn *websocket.Conn) *channel {
	return &channel{conn: conn}
}

func (c *channel) Send(ctx context.Context, msg message.Message) error {
	if err := wsjson.Write(ctx, c.conn, msg); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *channel) Receive(ctx context.Context) <-chan message.Message {
	messages := make(chan message.Message)
	go func() {
		defer func() {
			close(messages)
			_ = c.conn.Close(websocket.StatusNormalClosure, "context done or error")
		}()

		for {
			var msg message.Message
			err := wsjson.Read(ctx, c.conn, &msg)
			if err != nil {
				var closeErr websocket.CloseError
				if errors.As(err, &closeErr) || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					return
				}
				slog.Error("read error", "err", err)
			}
			select {
			case messages <- msg:
			case <-ctx.Done():
				return
			}

		}
	}()
	return messages
}

func (c *channel) Delete(_ context.Context) error { return nil }

func (c *channel) Close() error {
	if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
