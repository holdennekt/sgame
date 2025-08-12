package server

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

func NewQuestionStartedMessage() message.Message {
	return message.Message{Event: domain.QuestionStarted}
}

func HandleQuestionStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}

	time.AfterFunc(time.Until(room.CurrentQuestion.TimerEndsAt), func() {
		var question domain.Question

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			questionEnded :=
				newRoom.State != domain.ShowingQuestion ||
					!reflect.DeepEqual(newRoom.CurrentRoundQuestions, room.CurrentRoundQuestions)
			deadlineChanged := newRoom.CurrentQuestion != nil &&
				!room.CurrentQuestion.TimerEndsAt.Equal(newRoom.CurrentQuestion.TimerEndsAt)
			if questionEnded || deadlineChanged {
				return ErrDeferredFunctionCancelled
			}

			question = newRoom.CurrentQuestion.Question
			newRoom.EndQuestion()
			return nil
		})
		if err != nil {
			if err != ErrDeferredFunctionCancelled {
				log.Println(err)
			}
			return
		}

		questionEndedMessage := NewQuestionEndedMessage(question)
		if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
