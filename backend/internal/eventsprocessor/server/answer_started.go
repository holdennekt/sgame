package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type AnswerStartedMessage struct {
	domain.Question
	UserId string `json:"userId"`
}

func NewAnswerStartedMessage(question domain.Question, userId string) message.Message {
	payload, _ := json.Marshal(AnswerStartedMessage{
		Question: question,
		UserId:   userId,
	})
	return message.Message{Event: domain.AnswerStarted, Payload: payload}
}

func HandleAnswerStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, msg message.Message) error {
	var asp AnswerStartedMessage
	if err := json.Unmarshal(msg.Payload, &asp); err != nil {
		return err
	}
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(time.Until(room.AnsweringPlayer.TimerEndsAt), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			answerRequestEnded :=
				newRoom.State != domain.Answering ||
					newRoom.AnsweringPlayer.Id != asp.UserId ||
					!asp.Question.IsCurrent(newRoom)
			if answerRequestEnded {
				return ErrDeferredFunctionCancelled
			}

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
			questionEndedMessage := NewQuestionEndedMessage(asp.Question)
			if err := internalServer.Send(ctx, questionEndedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.ShowingQuestion:
			questionStartedMessage := NewQuestionStartedMessage(asp.Question)
			if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}

var ErrDeferredFunctionCancelled error = errors.New("deferred function is cancelled")
