package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/holdennekt/sgame/backend/pkg/metrics"
)

const (
	IDLE_ROOM_TTL       = 10 * time.Minute
	EXPIRE_GRACE_PERIOD = 1 * time.Second
)

type userDisconnectedPayload struct {
	UserId string `json:"userId"`
}

func NewUserDisconnectedMessage(userId string) message.Message {
	payload, _ := json.Marshal(userDisconnectedPayload{UserId: userId})
	return message.Message{Event: domain.UserDisconnected, Payload: payload}
}

func HandleUserDisconnectedMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, lobbyServer realtime.Channel, roomCache cache.Room, roomRepository repository.Room, roomId string, msg message.Message, idleRoomTTL time.Duration) error {
	var udp userDisconnectedPayload
	if err := json.Unmarshal(msg.Payload, &udp); err != nil {
		return err
	}

	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	isHostConnected := room.Host != nil && room.Host.IsConnected
	connectedPlayerIndex := slices.IndexFunc(room.Players, func(p domain.Player) bool {
		return p.IsConnected
	})
	if !isHostConnected && connectedPlayerIndex == -1 {
		if err := roomCache.Expire(ctx, roomId, idleRoomTTL); err != nil {
			return err
		}
		time.AfterFunc(idleRoomTTL+EXPIRE_GRACE_PERIOD, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			newRoom, err := roomCache.GetById(ctx, roomId)
			if newRoom != nil {
				return
			}
			if _, ok := err.(custerr.NotFoundErr); !ok {
				slog.Error("error", "err", err)
				return
			}
			metrics.RoomsActive.Dec()
			deletedRoomMsg := outgoing.NewRoomDeletedMessage(roomId)
			if err := lobbyServer.Send(ctx, deletedRoomMsg); err != nil {
				slog.Error("error", "err", err)
			}
			if err := server.Send(ctx, deletedRoomMsg); err != nil {
				slog.Error("error", "err", err)
			}
			if err := internalServer.Send(ctx, deletedRoomMsg); err != nil {
				slog.Error("error", "err", err)
			}
			if room.State != domain.WaitingForStart && room.State != domain.GameOver {
				if err := roomRepository.Create(ctx, room); err != nil {
					slog.Error("error", "err", err)
				}
			}
		})
	}
	return nil
}
