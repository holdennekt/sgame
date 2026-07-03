package server

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type QuestionEndedPayload struct {
	domain.Question
}

func NewQuestionEndedMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(QuestionEndedPayload{Question: question})
	return message.Message{Event: domain.QuestionEnded, Payload: payload}
}

func HandleQuestionEndedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, getAttachmentUrl func(key string) (string, error), roomId string, pack *domain.Pack, msg message.Message) error {
	var qep QuestionEndedPayload
	if err := json.Unmarshal(msg.Payload, &qep); err != nil {
		return err
	}

	var roundAdvanced bool
	var nextRoundStarted bool
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.State != domain.SelectingQuestion {
			return nil
		}

		if !room.AnyAvailableQuestions() {
			roundAdvanced = true
			nextRoundStarted = room.StartNextRegularRound(pack)
			if !nextRoundStarted {
				finalRoundStarted, err := room.StartFinalRound(pack, getAttachmentUrl)
				if err != nil {
					return err
				}
				if !finalRoundStarted {
					room.EndGame()
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	correctAnswerDemoMessage := outgoing.NewCorrectAnswerDemoMessage(qep.Question, getAttachmentUrl)
	if err := server.Send(ctx, correctAnswerDemoMessage); err != nil {
		return err
	}

	if !roundAdvanced {
		return nil
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.SelectingQuestion:
		if nextRoundStarted {
			roundStartedMessage := NewRoundStartedMessage()
			return internalServer.Send(ctx, roundStartedMessage)
		}
	case domain.FinalRoundBetting:
		bettingStartedMessage := NewFinalRoundBettingStartedMessage()
		return internalServer.Send(ctx, bettingStartedMessage)
	case domain.GameOver:
		gameEndedMessage := NewGameEndedMessage()
		return internalServer.Send(ctx, gameEndedMessage)
	}
	return nil
}
