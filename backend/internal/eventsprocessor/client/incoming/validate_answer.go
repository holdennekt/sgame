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

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.RevealingQuestion:
		revealingStartedMessage := serverevent.NewRevealingStartedMessage(question)
		if err := internalServer.Send(ctx, revealingStartedMessage); err != nil {
			return err
		}
	case domain.ShowingQuestion:
		questionStartedMessage := serverevent.NewQuestionStartedMessage(question)
		if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
			return err
		}
	case domain.SelectingQuestion:
		questionEndedMessage := serverevent.NewQuestionEndedMessage(question)
		if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
			return err
		}
	}
	return nil
}
