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

type ValidateAnswerPayload struct {
	IsCorrect bool `json:"isCorrect"`
}

func HandleValidateAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var vap ValidateAnswerPayload
	if err := json.Unmarshal(msg.Payload, &vap); err != nil {
		return err
	}
	var question domain.Question
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
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
	case domain.SelectingQuestion:
		questionEndedMessage := serverevent.NewQuestionEndedMessage(question)
		if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
			return err
		}
	case domain.ShowingQuestion:
		questionStartedMessage := serverevent.NewQuestionStartedMessage()
		if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
			return err
		}
	}
	return nil
}
