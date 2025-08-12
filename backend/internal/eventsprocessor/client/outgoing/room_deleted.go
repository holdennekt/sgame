package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type roomDeletedPayload struct {
	Id string `json:"id"`
}

func NewRoomDeletedMessage(id string) message.Message {
	payload, _ := json.Marshal(roomDeletedPayload{Id: id})
	return message.Message{
		Event:   domain.RoomDeleted,
		Payload: payload,
	}
}

func HandleRoomDeletedMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
