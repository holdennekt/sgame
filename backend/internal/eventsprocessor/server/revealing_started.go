package server

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

func NewRevealingStartedMessage() message.Message {
	return message.Message{Event: domain.RevealingStarted}
}

func HandleRevealingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}

	time.AfterFunc(time.Until(room.CurrentQuestion.TimerStartsAt), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			questionEnded :=
				newRoom.State != domain.RevealingQuestion ||
					!reflect.DeepEqual(newRoom.CurrentRoundQuestions, room.CurrentRoundQuestions)
			deadlineChanged := newRoom.CurrentQuestion != nil &&
				!room.CurrentQuestion.TimerStartsAt.Equal(newRoom.CurrentQuestion.TimerStartsAt)
			if questionEnded || deadlineChanged {
				return ErrDeferredFunctionCancelled
			}

			newRoom.StartRegularQuestion()
			return nil
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

		questionStartedMessage := NewQuestionStartedMessage()
		if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
