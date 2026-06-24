package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
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
	spectatorCount, _ := roomCache.GetSpectatorCount(ctx, roomUpdatedPayload.Id)
	payload, _ := json.Marshal(room.GetProjection(user.Id, spectatorCount))
	return client.Send(ctx, message.Message{Event: msg.Event, Payload: payload})
}

func HandleLobbyRoomUpdatedMessage(ctx context.Context, roomCache cache.Room, client realtime.Channel, msg message.Message) error {
	var roomUpdatedPayload roomUpdatedPayload
	if err := json.Unmarshal(msg.Payload, &roomUpdatedPayload); err != nil {
		return err
	}
	room, _ := roomCache.GetById(ctx, roomUpdatedPayload.Id)
	payload, _ := json.Marshal(domain.NewRoomLobby(room))
	return client.Send(ctx, message.Message{Event: msg.Event, Payload: payload})
}
