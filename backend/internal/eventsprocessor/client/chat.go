package client

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type chatPayload struct {
	From domain.User `json:"from"`
	incomingChatPayload
}

type incomingChatPayload struct {
	Text string `json:"text"`
}

func NewChatMessage(user domain.User, msg message.Message) (message.Message, error) {
	var incomingPayload incomingChatPayload
	if err := json.Unmarshal(msg.Payload, &incomingPayload); err != nil {
		return message.Message{}, err
	}
	payload, _ := json.Marshal(chatPayload{
		incomingChatPayload: incomingPayload,
		From:                user,
	})
	return message.Message{
		Event:   domain.Chat,
		Payload: payload,
	}, nil
}

func NewSystemChatMessage(text string) message.Message {
	payload, _ := json.Marshal(chatPayload{
		incomingChatPayload: incomingChatPayload{Text: text},
	})
	return message.Message{
		Event:   domain.Chat,
		Payload: payload,
	}
}

func HandleClientChatMessage(ctx context.Context, server realtime.Channel, user domain.User, msg message.Message) error {
	chatMessage, err := NewChatMessage(user, msg)
	if err != nil {
		return err
	}
	return server.Send(ctx, chatMessage)
}

func HandleServerChatMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
