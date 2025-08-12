package server

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type QuestionEndedPayload struct {
	Question domain.Question `json:"question"`
}

func NewQuestionEndedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(QuestionEndedPayload{Question: question})
	return message.Message{Event: domain.QuestionEnded, Payload: payload}
}

func HandleQuestionEndedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, pack *domain.Pack, msg message.Message) error {
	var qep QuestionEndedPayload
	if err := json.Unmarshal(msg.Payload, &qep); err != nil {
		return err
	}

	var nextRoundStarted bool
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		if !room.AnyAvailableQuestions() {
			nextRoundStarted = room.StartNextRegularRound(pack)
			if !nextRoundStarted {
				if !room.StartFinalRound(pack) {
					room.EndGame()
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	correctAnswerDemoMessage := outgoing.NewCorrectAnswerDemoMessage(qep.Question)
	if err := server.Send(ctx, correctAnswerDemoMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.SelectingQuestion:
		if nextRoundStarted {
			roundStartedMessage := NewRoundStartedMessage()
			if err := internalServer.Send(ctx, roundStartedMessage); err != nil {
				return err
			}
		}
	case domain.FinalRoundBetting:
		bettingStartedMessage := NewFinalRoundBettingStartedMessage()
		if err := internalServer.Send(ctx, bettingStartedMessage); err != nil {
			return err
		}
	case domain.GameOver:
		gameEndedMessage := NewGameEndedMessage()
		if err := internalServer.Send(ctx, gameEndedMessage); err != nil {
			return err
		}
	}
	return nil
}
