package incoming

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func HandleUnpauseMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(r *domain.Room) error {
		return r.Unpause(user.Id)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	return sendRescheduleMessage(ctx, internalServer, newRoom)
}
