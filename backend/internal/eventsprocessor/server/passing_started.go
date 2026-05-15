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

const TimeToPass = 30 * time.Second

type PassingStartedPayload struct {
	domain.Question
}

func NewPassingStartedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(PassingStartedPayload{Question: question})
	return message.Message{Event: domain.PassingStarted, Payload: payload}
}

func HandlePassingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, msg message.Message) error {
	var psp PassingStartedPayload
	if err := json.Unmarshal(msg.Payload, &psp); err != nil {
		return err
	}
	time.AfterFunc(TimeToPass, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			if newRoom.State != domain.Passing || !psp.Question.IsCurrent(newRoom) {
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

		answerStartedMessage := NewAnswerStartedMessage(psp.Question, newerRoom.AnsweringPlayer.Id)
		if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
