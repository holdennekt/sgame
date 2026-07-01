package incoming

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/domain"
	clientevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type BanPlayerPayload struct {
	PlayerId string `json:"playerId"`
}

func HandleBanPlayerMessage(ctx context.Context, server realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var bpp BanPlayerPayload
	if err := json.Unmarshal(msg.Payload, &bpp); err != nil {
		return err
	}

	var targetName string
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		playerIdx := room.UsersPlayerIndex(bpp.PlayerId)
		if playerIdx != -1 {
			targetName = room.Players[playerIdx].Name
		}
		return room.BanPlayer(user.Id, bpp.PlayerId)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	chatMsg := clientevent.NewSystemChatMessage(fmt.Sprintf("%s banned %s from the room", user.Name, targetName))
	return server.Send(ctx, chatMsg)
}
