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

type RevealingStartedPayload struct {
	domain.Question
}

func NewRevealingStartedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(RevealingStartedPayload{Question: question})
	return message.Message{Event: domain.RevealingStarted, Payload: payload}
}

func HandleRevealingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, msg message.Message) error {
	var rsp RevealingStartedPayload
	if err := json.Unmarshal(msg.Payload, &rsp); err != nil {
		return err
	}
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
					!rsp.Question.IsCurrent(newRoom)
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

		questionStartedMessage := NewQuestionStartedMessage(rsp.Question)
		if err := internalServer.Send(ctx, questionStartedMessage); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
