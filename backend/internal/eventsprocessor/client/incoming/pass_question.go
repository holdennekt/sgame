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

type PassQuestionPayload struct {
	PassTo string `json:"passTo"`
}

func HandlePassQuestionMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var pqp PassQuestionPayload
	if err := json.Unmarshal(msg.Payload, &pqp); err != nil {
		return err
	}
	_, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.PassQuestion(user.Id, pqp.PassTo)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	answerStartedMessage := serverevent.NewAnswerStartedMessage()
	if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
		return err
	}
	return nil
}
