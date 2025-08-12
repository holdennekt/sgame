package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type roomUpdatedPayload struct {
	Id string `json:"id"`
}

func NewRoomUpdatedMessage(roomId string) message.Message {
	payload, _ := json.Marshal(roomUpdatedPayload{Id: roomId})
	return message.Message{Event: domain.RoomUpdated, Payload: payload}
}

func HandleRoomUpdatedMessage(ctx context.Context, roomCache cache.Room, client realtime.Channel, user domain.User, msg message.Message) error {
	var roomUpdatedPayload roomUpdatedPayload
	if err := json.Unmarshal(msg.Payload, &roomUpdatedPayload); err != nil {
		return err
	}
	room, _ := roomCache.GetById(ctx, roomUpdatedPayload.Id)
	payload, _ := json.Marshal(room.GetProjection(user.Id))
	return client.Send(ctx, message.Message{Event: msg.Event, Payload: payload})
}
