package incoming

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type RemoveFinalRoundCategoryPayload struct {
	Category string `json:"category"`
}

func HandleRemoveFinalRoundCategoryMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, pack *domain.Pack, msg message.Message) error {
	var rfrcp RemoveFinalRoundCategoryPayload
	if err := json.Unmarshal(msg.Payload, &rfrcp); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.RemoveFinalRoundCategory(pack, user.Id, rfrcp.Category)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	if newRoom.State == domain.FinalRoundBetting {
		bettingStartedMessage := serverevent.NewFinalRoundBettingStartedMessage()
		if err := internalServer.Send(ctx, bettingStartedMessage); err != nil {
			return err
		}
	}
	return nil
}
