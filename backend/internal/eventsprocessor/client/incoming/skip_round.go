package incoming

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/domain"
	clientevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
)

func HandleSkipRoundMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, getAttachmentUrl func(string) (string, error), roomId string, user domain.User, pack *domain.Pack) error {
	var nextRoundStarted bool
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if err := room.SkipRound(user.Id); err != nil {
			return err
		}
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
		return nil
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}
	if err := server.Send(ctx, clientevent.NewSystemChatMessage(fmt.Sprintf("%s skipped the round", user.Name))); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.SelectingQuestion:
		if nextRoundStarted {
			roundStartedMessage := serverevent.NewRoundStartedMessage()
			return internalServer.Send(ctx, roundStartedMessage)
		}
	case domain.FinalRoundBetting:
		bettingStartedMessage := serverevent.NewFinalRoundBettingStartedMessage()
		return internalServer.Send(ctx, bettingStartedMessage)
	case domain.GameOver:
		gameEndedMessage := serverevent.NewGameEndedMessage()
		return internalServer.Send(ctx, gameEndedMessage)
	}
	return nil
}
