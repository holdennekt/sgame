package ws

import (
	"context"
	"errors"
	"log"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/holdennekt/sgame/internal/message"
	"github.com/holdennekt/sgame/pkg/custerr"
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

func (c *channel) Recieve(ctx context.Context) <-chan message.Message {
	messages := make(chan message.Message)
	go func() {
		defer func() {
			close(messages)
			c.conn.Close(websocket.StatusNormalClosure, "context done or error")
		}()

		for {
			var msg message.Message
			err := wsjson.Read(ctx, c.conn, &msg)
			if err != nil {
				var closeErr websocket.CloseError
				if !errors.As(err, &closeErr) {
					log.Println("read error:", err)
				}
				return
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

func (c *channel) Close() error {
	if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
