package incoming

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type SelectQuestionPayload struct {
	Category string `json:"category"`
	Index    int    `json:"index"`
}

func HandleSelectQuestionMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, pack *domain.Pack, msg message.Message) error {
	var qsp SelectQuestionPayload
	if err := json.Unmarshal(msg.Payload, &qsp); err != nil {
		return err
	}
	newRoom, err := roomCache.SafeSet(ctx, roomId, func(room *domain.Room) error {
		return room.SelectQuestion(user.Id, pack, qsp.Category, qsp.Index)
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
	if err := server.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}

	switch newRoom.State {
	case domain.RevealingQuestion:
		revealingStartedMessage := serverevent.NewRevealingStartedMessage()
		if err := internalServer.Send(ctx, revealingStartedMessage); err != nil {
			return err
		}
	case domain.Answering:
		answerStartedMessage := serverevent.NewAnswerStartedMessage()
		if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
			return err
		}
	case domain.Passing:
		passingStartedMessage := serverevent.NewPassingStartedMessage()
		if err := internalServer.Send(ctx, passingStartedMessage); err != nil {
			return err
		}
	case domain.Betting:
		bettingStartedMessage := serverevent.NewBettingStartedMessage()
		if err := internalServer.Send(ctx, bettingStartedMessage); err != nil {
			return err
		}
	}
	return nil
}
