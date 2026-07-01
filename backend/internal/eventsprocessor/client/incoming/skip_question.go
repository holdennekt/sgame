package incoming

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/domain"
	clientevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func HandleSkipQuestionMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var question domain.Question
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
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
	if err := server.Send(ctx, clientevent.NewSystemChatMessage(fmt.Sprintf("%s skipped the question", user.Name))); err != nil {
		return err
	}

	questionEndedMessage := serverevent.NewQuestionEndedMessage(question)
	return internalServer.Send(ctx, questionEndedMessage)
}
