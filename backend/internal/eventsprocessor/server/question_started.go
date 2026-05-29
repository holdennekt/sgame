package server

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type QuestionStartedPayload struct {
	domain.Question
}

func NewQuestionStartedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(QuestionStartedPayload{Question: question})
	return message.Message{Event: domain.QuestionStarted, Payload: payload}
}

func HandleQuestionStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, msg message.Message) error {
	var qsp QuestionStartedPayload
	if err := json.Unmarshal(msg.Payload, &qsp); err != nil {
		return err
	}
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}

	time.AfterFunc(time.Until(room.CurrentQuestion.TimerEndsAt), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := roomCache.SafeUpdate(ctx, roomId, func(newRoom *domain.Room) error {
			questionEnded :=
				newRoom.State != domain.ShowingQuestion ||
					!qsp.Question.IsCurrent(newRoom)
			deadlineChanged := newRoom.CurrentQuestion != nil &&
				!room.CurrentQuestion.TimerEndsAt.Equal(newRoom.CurrentQuestion.TimerEndsAt)
			if questionEnded || deadlineChanged || newRoom.PausedState.Paused {
				return ErrDeferredFunctionCancelled
			}

			newRoom.EndQuestion()
			return nil
		})
		if err != nil {
			if err != ErrDeferredFunctionCancelled {
				log.Println(err)
			}
			return
		}

		questionEndedMessage := NewQuestionEndedMessage(qsp.Question)
		if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
