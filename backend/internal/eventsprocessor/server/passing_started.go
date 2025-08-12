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

const TimeToPass = 30 * time.Second

func NewPassingStartedMessage() message.Message {
	return message.Message{Event: domain.PassingStarted}
}

func HandlePassingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(TimeToPass, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			questionEnded :=
				!reflect.DeepEqual(newRoom.CurrentRoundQuestions, room.CurrentRoundQuestions)
			if newRoom.State != domain.Passing || questionEnded {
				return ErrDeferredFunctionCancelled
			}

			newRoom.PassQuestionAuto()
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

		answerStartedMessage := NewAnswerStartedMessage()
		if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
