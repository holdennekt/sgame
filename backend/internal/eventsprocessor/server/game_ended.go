package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/metrics"
)

func NewGameEndedMessage() message.Message {
	return message.Message{Event: domain.GameEnded}
}

func HandleGameEndedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, lobbyServer realtime.Channel, roomCache cache.Room, roomRepository repository.Room, roomId string, idleRoomTTL time.Duration) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	now := time.Now()
	room.FinishedAt = &now
	if err := roomRepository.Create(ctx, room); err != nil {
		return err
	}
	time.AfterFunc(idleRoomTTL, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := roomCache.Delete(ctx, roomId); err != nil {
			slog.Error("error", "err", err)
			return
		}
		metrics.RoomsActive.Dec()
		deletedRoomMsg := outgoing.NewRoomDeletedMessage(roomId)
		if err := lobbyServer.Send(ctx, deletedRoomMsg); err != nil {
			slog.Error("error", "err", err)
			return
		}
		if err := server.Send(ctx, deletedRoomMsg); err != nil {
			slog.Error("error", "err", err)
			return
		}
		if err := internalServer.Send(ctx, deletedRoomMsg); err != nil {
			slog.Error("error", "err", err)
			return
		}
	})
	return nil
}
