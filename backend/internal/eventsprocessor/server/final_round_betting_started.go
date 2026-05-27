package server

import (
	"context"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func NewFinalRoundBettingStartedMessage() message.Message {
	return message.Message{Event: domain.FinalRoundBettingStarted}
}

func HandleFinalRoundBettingStartedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	time.AfterFunc(time.Until(*room.FinalRoundState.BettingEndsAt), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newerRoom, err := roomCache.SafeSet(ctx, roomId, func(newRoom *domain.Room) error {
			if newRoom.State != domain.FinalRoundBetting {
				return ErrDeferredFunctionCancelled
			}
			deadlineChanged := newRoom.FinalRoundState != nil &&
				newRoom.FinalRoundState.BettingEndsAt != nil &&
				!room.FinalRoundState.BettingEndsAt.Equal(*newRoom.FinalRoundState.BettingEndsAt)
			if deadlineChanged || newRoom.PausedState.Paused {
				return ErrDeferredFunctionCancelled
			}
			newRoom.PlaceFinalRoundBetsAuto()
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
		case domain.ShowingFinalRoundQuestion:
			finalRoundQuestionStartedMessage := NewFinalRoundQuestionStartedMessage()
			if err := internalServer.Send(ctx, finalRoundQuestionStartedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.GameOver:
			gameEndedMessage := NewGameEndedMessage()
			if err := internalServer.Send(ctx, gameEndedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}
