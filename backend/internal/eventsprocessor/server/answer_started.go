package server

import (
	"context"
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

func NewAnswerStartedMessage() message.Message {
	return message.Message{Event: domain.AnswerStarted}
}

func HandleAnswerStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(time.Until(room.AnsweringPlayer.TimerEndsAt), func() {
		var question domain.Question

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			answerRequestEnded :=
				newRoom.State != domain.Answering ||
					newRoom.AnsweringPlayer.Id != room.AnsweringPlayer.Id ||
					!reflect.DeepEqual(newRoom.CurrentRoundQuestions, room.CurrentRoundQuestions)
			if answerRequestEnded {
				return ErrDeferredFunctionCancelled
			}

			question = newRoom.CurrentQuestion.Question
			return newRoom.ValidateAnswer(domain.SYSTEM, false)
		})
		if err != nil {
			if err != ErrDeferredFunctionCancelled {
				log.Println(err)
			}
			return
		}

		roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
		if err := server.Send(ctx, roomUpdatedMessage); err != nil {
			log.Println(err)
			return
		}

		switch newerRoom.State {
		case domain.SelectingQuestion:
			questionEndedMessage := NewQuestionEndedMessage(question)
			if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.ShowingQuestion:
			questionStartedMessage := NewQuestionStartedMessage()
			if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}

var ErrDeferredFunctionCancelled error = errors.New("deferred function is cancelled")
