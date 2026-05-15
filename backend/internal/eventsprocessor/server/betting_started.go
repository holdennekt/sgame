package server

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

const TimeToBet = 30 * time.Second

type BettingStartedPayload struct {
	domain.Question
}

func NewBettingStartedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(BettingStartedPayload{Question: question})
	return message.Message{Event: domain.BettingStarted, Payload: payload}
}

func HandleBettingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, msg message.Message) error {
	var bsp BettingStartedPayload
	if err := json.Unmarshal(msg.Payload, &bsp); err != nil {
		return err
	}
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(TimeToBet, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			if newRoom.State != domain.Betting || !bsp.Question.IsCurrent(newRoom) {
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
			answerStartedMessage := NewAnswerStartedMessage(bsp.Question, newerRoom.AnsweringPlayer.Id)
			if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}
