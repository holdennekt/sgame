package incoming

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type SubmitFinalRoundAnswerPayload struct {
	Answer string `json:"answer"`
}

func HandleSubmitFinalRoundAnswerMessage(ctx context.Context, server realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var afrap SubmitFinalRoundAnswerPayload
	if err := json.Unmarshal(msg.Payload, &afrap); err != nil {
		return err
	}
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		return room.SubmitFinalRoundAnswer(user.Id, afrap.Answer)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	return nil
}
