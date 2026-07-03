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

func HandleStartAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var question domain.Question
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.CurrentQuestion != nil {
			question = room.CurrentQuestion.Question
		}
		return room.StartAnswer(user.Id)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.RevealingQuestion:
		return server.Send(ctx, outgoing.NewFalseStartMessage(user))
	case domain.SelectingQuestion:
		return internalServer.Send(ctx, serverevent.NewQuestionEndedMessage(question))
	case domain.Answering:
		answerStartedMessage := serverevent.NewAnswerStartedMessage(
			newRoom.CurrentQuestion.Question,
			newRoom.AnsweringPlayer.Id,
		)
		return internalServer.Send(ctx, answerStartedMessage)
	}
	return nil
}
