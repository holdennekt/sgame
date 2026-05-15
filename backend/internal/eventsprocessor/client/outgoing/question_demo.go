package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

const QuestionDemoDuration = 5

type QuestionDemoPayload struct {
	Category string              `json:"category"`
	Value    int                 `json:"value"`
	Type     domain.QuestionType `json:"type"`
	Duration int                 `json:"duration"`
}

func NewQuestionDemoMessage(question domain.Question) message.Message {
	payload, _ := json.Marshal(QuestionDemoPayload{
		Category: question.Category,
		Value:    question.Value,
		Type:     question.Type,
		Duration: QuestionDemoDuration,
	})
	return message.Message{
		Event:   domain.QuestionDemo,
		Payload: payload,
	}
}

func HandleQuestionDemoMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
