package incoming

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

func HandlePauseMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	room, err := roomCache.SafeSet(ctx, roomId, func(r *domain.Room) error {
		return r.Pause(user.Id)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	pausedAt := *room.PausedState.PausedAt
	time.AfterFunc(domain.MaxPauseDuration, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		newRoom, err := roomCache.SafeSet(ctx, roomId, func(r *domain.Room) error {
			return r.UnpauseSystem(pausedAt)
		})
		if err != nil {
			var conflictErr custerr.ConflictErr
			if !errors.As(err, &conflictErr) {
				log.Println(err)
			}
			return
		}

		if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
			log.Println(err)
			return
		}

		if err := sendRescheduleMessage(ctx, internalServer, newRoom); err != nil {
			log.Println(err)
		}
	})
	return nil
}

func sendRescheduleMessage(ctx context.Context, internalServer realtime.Channel, room *domain.Room) error {
	switch room.State {
	case domain.RevealingQuestion:
		return internalServer.Send(ctx, serverevent.NewRevealingStartedMessage(room.CurrentQuestion.Question))
	case domain.ShowingQuestion:
		return internalServer.Send(ctx, serverevent.NewQuestionStartedMessage(room.CurrentQuestion.Question))
	case domain.Answering:
		return internalServer.Send(ctx, serverevent.NewAnswerStartedMessage(room.CurrentQuestion.Question, room.AnsweringPlayer.Id))
	case domain.Betting:
		return internalServer.Send(ctx, serverevent.NewBettingStartedMessage(room.CurrentQuestion.Question))
	case domain.Passing:
		return internalServer.Send(ctx, serverevent.NewPassingStartedMessage(room.CurrentQuestion.Question))
	case domain.FinalRoundBetting:
		return internalServer.Send(ctx, serverevent.NewFinalRoundBettingStartedMessage())
	case domain.ShowingFinalRoundQuestion:
		return internalServer.Send(ctx, serverevent.NewFinalRoundQuestionStartedMessage())
	}
	return nil
}
