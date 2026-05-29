package incoming

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type ChangeScorePayload struct {
	PlayerId string `json:"playerId"`
	Score    int    `json:"score"`
}

func HandleChangeScoreMessage(ctx context.Context, server realtime.Channel, roomCache cache.Room, roomId string, user domain.User, msg message.Message) error {
	var csp ChangeScorePayload
	if err := json.Unmarshal(msg.Payload, &csp); err != nil {
		return err
	}
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		return room.ChangeScore(user.Id, csp.PlayerId, csp.Score)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	return server.Send(ctx, roomUpdatedMessage)
}
