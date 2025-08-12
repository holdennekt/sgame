package incoming

import (
	"context"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

func HandleSubmitAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	_, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.SubmitAnswer(user.Id)
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
