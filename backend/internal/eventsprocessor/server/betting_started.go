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

const TimeToBet = 30 * time.Second

func NewBettingStartedMessage() message.Message {
	return message.Message{Event: domain.BettingStarted}
}

func HandleBettingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(TimeToBet, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			questionEnded :=
				!reflect.DeepEqual(newRoom.CurrentRoundQuestions, room.CurrentRoundQuestions)
			if newRoom.State != domain.Betting || questionEnded {
				return ErrDeferredFunctionCancelled
			}

			newRoom.PlaceBetsAuto()
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

		switch newerRoom.State {
		case domain.SelectingQuestion:
			questionEndedMessage := NewQuestionEndedMessage(room.CurrentQuestion.Question)
			if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.ShowingQuestion:
			answerStartedMessage := NewAnswerStartedMessage()
			if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}
