package incoming

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func HandleSkipQuestionMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var question domain.Question
	_, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		if room.CurrentQuestion != nil {
			question = room.CurrentQuestion.Question
		}
		return room.SkipQuestion(user.Id)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	questionEndedMessage := serverevent.NewQuestionEndedMessage(question)
	return internalServer.Send(ctx, questionEndedMessage)
}
