package incoming

import (
	"context"
	"errors"
	"log"
	"slices"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func HandleStartGameMessage(ctx context.Context, lobbyServer realtime.Channel, roomServer realtime.Channel, roomInternalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, pack *domain.Pack, msg message.Message) error {
	_, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		anyConnectedPlayer := slices.ContainsFunc(room.Players, func(p domain.Player) bool {
			return p.IsConnected
		})
		if !room.IsUserHost(user.Id) || !anyConnectedPlayer {
			return errors.New("not allowed to start game")
		}
		room.StartGame(pack)
		return nil
	})
	if err != nil {
		return err
	}

	roundStartedMessage := serverevent.NewRoundStartedMessage()
	if err := roomInternalServer.Send(ctx, roundStartedMessage); err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := lobbyServer.Send(ctx, roomUpdatedMessage); err != nil {
		log.Println(err)
	}

	return roomServer.Send(ctx, roomUpdatedMessage)
}
