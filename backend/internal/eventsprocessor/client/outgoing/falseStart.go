package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type falseStartPayload struct {
	User domain.User `json:"user"`
}

func NewFalseStartMessage(user domain.User) message.Message {
	payload, _ := json.Marshal(falseStartPayload{User: user})
	return message.Message{
		Event:   domain.FalseStart,
		Payload: payload,
	}
}

func HandleFalseStartMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
