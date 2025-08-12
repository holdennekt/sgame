package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type RoundDemoPayload struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
}

func NewRoundDemoMessage(round domain.Round) message.Message {
	categories := make([]string, 0)
	for _, category := range round.Categories {
		categories = append(categories, category.Name)
	}
	payload, _ := json.Marshal(RoundDemoPayload{
		Name:       round.Name,
		Categories: categories,
	})
	return message.Message{
		Event:   domain.RoundDemo,
		Payload: payload,
	}
}

func HandleRoundDemoMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
