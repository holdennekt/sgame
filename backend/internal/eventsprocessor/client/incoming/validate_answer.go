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
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type ValidateAnswerPayload struct {
	IsCorrect bool `json:"isCorrect"`
}

func HandleValidateAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var vap ValidateAnswerPayload
	if err := json.Unmarshal(msg.Payload, &vap); err != nil {
		return err
	}
	var question domain.Question
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.CurrentQuestion == nil {
			return custerr.NewConflictErr("can not validate answer now")
		}
		question = room.CurrentQuestion.Question
		return room.ValidateAnswer(user.Id, vap.IsCorrect)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	return handlePostValidate(ctx, internalServer, newRoom, question)
}

func handlePostValidate(ctx context.Context, internalServer realtime.Channel, newRoom *domain.Room, question domain.Question) error {
	switch newRoom.State {
	case domain.RevealingQuestion:
		return internalServer.Send(ctx, serverevent.NewRevealingStartedMessage(question))
	case domain.ShowingQuestion:
		return internalServer.Send(ctx, serverevent.NewQuestionStartedMessage(question))
	case domain.SelectingQuestion:
		return internalServer.Send(ctx, serverevent.NewQuestionEndedMessage(question))
	}
	return nil
}
