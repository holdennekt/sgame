package server

import (
	"context"
	"log"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/message"
)

func NewGameEndedMessage() message.Message {
	return message.Message{Event: domain.GameEnded}
}

func HandleGameEndedMessage(ctx context.Context, server realtime.Channel, lobbyServer realtime.Channel, roomCache cache.Room, roomRepository repository.Room, roomId string) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	if err := roomRepository.Create(ctx, room); err != nil {
		return err
	}
	time.AfterFunc(IDLE_ROOM_TTL, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := roomCache.Delete(ctx, roomId); err != nil {
			log.Println(err)
			return
		}
		deletedRoomMsg := outgoing.NewRoomDeletedMessage(roomId)
		if err := lobbyServer.Send(ctx, deletedRoomMsg); err != nil {
			log.Println(err)
			return
		}
		if err := server.Send(ctx, deletedRoomMsg); err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
