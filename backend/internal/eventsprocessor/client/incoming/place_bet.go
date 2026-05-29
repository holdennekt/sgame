package incoming

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type PlaceBetPayload struct {
	Amount int `json:"amount"`
}

func HandlePlaceBetMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var pbp PlaceBetPayload
	if err := json.Unmarshal(msg.Payload, &pbp); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		return room.PlaceBet(user.Id, pbp.Amount)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	if newRoom.State == domain.Answering {
		answerStartedMessage := serverevent.NewAnswerStartedMessage(
			newRoom.CurrentQuestion.Question,
			newRoom.AnsweringPlayer.Id,
		)
		if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
			return err
		}
	}
	return nil
}
