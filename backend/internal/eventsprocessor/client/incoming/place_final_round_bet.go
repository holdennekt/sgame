package incoming

import (
	"context"
	"encoding/json"
	"log"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

func HandlePlaceFinalRoundBetMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var pbp PlaceBetPayload
	if err := json.Unmarshal(msg.Payload, &pbp); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.PlaceFinalRoundBet(user.Id, pbp.Amount)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.ShowingFinalRoundQuestion:
		finalRoundQuestionStartedMessage := serverevent.NewFinalRoundQuestionStartedMessage()
		if err := internalServer.Send(ctx, finalRoundQuestionStartedMessage); err != nil {
			log.Println(err)
		}
	case domain.GameOver:
		gameEndedMessage := serverevent.NewGameEndedMessage()
		if err := internalServer.Send(ctx, gameEndedMessage); err != nil {
			log.Println(err)
		}
	}
	return nil
}
