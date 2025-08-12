package server

import (
	"context"
	"encoding/json"
	"log"
	"slices"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/message"
	"github.com/holdennekt/sgame/pkg/custerr"
)

const (
	IDLE_ROOM_TTL       = 5 * time.Minute
	EXPIRE_GRACE_PERIOD = 1 * time.Second
)

type userDisconnectedPayload struct {
	UserId string `json:"userId"`
}

func NewUserDisconnectedMessage(userId string) message.Message {
	payload, _ := json.Marshal(userDisconnectedPayload{UserId: userId})
	return message.Message{Event: domain.UserDisconnected, Payload: payload}
}

func HandleUserDisconnectedMessage(ctx context.Context, server realtime.Channel, lobbyServer realtime.Channel, roomCache cache.Room, roomRepository repository.Room, roomId string, msg message.Message) error {
	var udp userDisconnectedPayload
	if err := json.Unmarshal(msg.Payload, &udp); err != nil {
		return err
	}

	// Pause if needed
	// newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
	// 	hostConnected := room.Host != nil && room.Host.IsConnected
	// 	connectedPlayerIndex := slices.IndexFunc(room.Players, func(p domain.Player) bool {
	// 		return p.IsConnected
	// 	})
	// 	playerAnswering := room.State == domain.Answering && room.AnsweringPlayer.Id == udp.UserId
	// 	if (!hostConnected || connectedPlayerIndex == -1 || playerAnswering) &&
	// 		!room.PausedState.Paused {
	// 		room.PausedState.Paused = true
	// 		now := time.Now()
	// 		room.PausedState.PausedAt = &now
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	return err
	// }

	// Set expiration time if needed
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	isHostConnected := room.Host != nil && room.Host.IsConnected
	connectedPlayerIndex := slices.IndexFunc(room.Players, func(p domain.Player) bool {
		return p.IsConnected
	})
	if !isHostConnected && connectedPlayerIndex == -1 {
		if err := roomCache.Expire(ctx, roomId, IDLE_ROOM_TTL); err != nil {
			return err
		}
		time.AfterFunc(IDLE_ROOM_TTL+EXPIRE_GRACE_PERIOD, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			newRoom, err := roomCache.GetById(ctx, roomId)
			if newRoom != nil {
				return
			}
			if _, ok := err.(custerr.NotFoundErr); !ok {
				log.Println(err)
				return
			}
			deletedRoomMsg := outgoing.NewRoomDeletedMessage(roomId)
			if err := lobbyServer.Send(ctx, deletedRoomMsg); err != nil {
				log.Println(err)
			}
			if err := server.Send(ctx, deletedRoomMsg); err != nil {
				log.Println(err)
			}
			if room.State != domain.WaitingForStart && room.State != domain.GameOver {
				if err := roomRepository.Create(ctx, room); err != nil {
					log.Println(err)
				}
			}
		})
	}
	return nil
}
