package server

import (
	"context"
	"slices"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func NewRoundStartedMessage() message.Message {
	return message.Message{Event: domain.RoundStarted}
}

func HandleRoundStartedMessage(ctx context.Context, server realtime.Channel, roomCache cache.Room, roomId string, pack *domain.Pack) error {
	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	currentRoundIndex := slices.IndexFunc(pack.Rounds, func(round domain.Round) bool {
		return *room.CurrentRoundName == round.Name
	})
	roundDemoMessage := outgoing.NewRoundDemoMessage(pack.Rounds[currentRoundIndex])
	if err := server.Send(ctx, roundDemoMessage); err != nil {
		return err
	}
	return nil
}
