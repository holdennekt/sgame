package incoming

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type SubmitFinalRoundAnswerPayload struct {
	Answer string `json:"answer"`
}

func HandleSubmitFinalRoundAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, getAttachmentUrl func(key string) (string, error), roomId string, user domain.User, validator ivalidator.AnswerValidator, cfg *config.Config, msg message.Message) error {
	var afrap SubmitFinalRoundAnswerPayload
	if err := json.Unmarshal(msg.Payload, &afrap); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		return room.SubmitFinalRoundAnswer(user.Id, afrap.Answer)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	if newRoom.State == domain.ValidatingFinalRoundAnswers && newRoom.Options.AIHost {
		go func() {
			for newRoom.State == domain.ValidatingFinalRoundAnswers {
				playerId := *newRoom.CurrentPlayer
				answer := newRoom.FinalRoundState.PlayersAnswers[playerId]
				iQuestion := buildFinalRoundValidatorQuestion(*newRoom.FinalRoundState.Question)
				isCorrect := resolveIsCorrect(validator, cfg, roomId, iQuestion, answer)

				finalCtx, finalCancel := context.WithTimeout(context.Background(), 5*time.Second)
				newerRoom, err := roomCache.SafeUpdate(finalCtx, roomId, func(r *domain.Room) error {
					validatingEnded := r.State != domain.ValidatingFinalRoundAnswers || r.CurrentPlayer == nil || *r.CurrentPlayer != playerId
					if validatingEnded {
						return serverevent.ErrDeferredFunctionCancelled
					}
					return r.ValidateFinalRoundAnswer(domain.SYSTEM, isCorrect)
				})
				finalCancel()
				if err != nil {
					if !errors.Is(err, serverevent.ErrDeferredFunctionCancelled) {
						slog.Error("error applying AI validation result for final round", "err", err, "room_id", roomId)
					}
					return
				}

				if err := server.Send(context.Background(), outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
					slog.Error("error sending room_updated after final round AI validation", "err", err)
					return
				}

				if newerRoom.State == domain.GameOver {
					if err := internalServer.Send(context.Background(), serverevent.NewGameEndedMessage()); err != nil {
						slog.Error("error", "err", err)
					}
					correctAnswerDemoMessage := outgoing.NewCorrectAnswerDemoMessage(newerRoom.FinalRoundState.Question, getAttachmentUrl)
					if err := server.Send(context.Background(), correctAnswerDemoMessage); err != nil {
						slog.Error("error", "err", err)
					}
					return
				}

				newRoom = newerRoom
			}
		}()
	}

	return nil
}
