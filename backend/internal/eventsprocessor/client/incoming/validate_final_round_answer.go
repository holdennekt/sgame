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

func HandleValidateFinalRoundAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var vap ValidateAnswerPayload
	if err := json.Unmarshal(msg.Payload, &vap); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.ValidateFinalRoundAnswer(user.Id, vap.IsCorrect)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	if newRoom.State == domain.GameOver {
		gameEndedMessage := serverevent.NewGameEndedMessage()
		if err := internalServer.Send(ctx, gameEndedMessage); err != nil {
			log.Println(err)
		}

		correctAnswerDemoMessage := outgoing.NewCorrectAnswerDemoMessage(newRoom.FinalRoundState.Question)
		if err := server.Send(ctx, correctAnswerDemoMessage); err != nil {
			log.Println(err)
		}
	}
	return nil
}
