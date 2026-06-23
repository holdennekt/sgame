package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type QuestionDemoPayload struct {
	Category        string              `json:"category"`
	CategoryComment *string             `json:"categoryComment"`
	Value           int                 `json:"value"`
	Type            domain.QuestionType `json:"type"`
	Duration        int                 `json:"duration"`
}

func NewQuestionDemoMessage(question domain.Question, categoryComment *string, demoDuration int) message.Message {
	payload, _ := json.Marshal(QuestionDemoPayload{
		Category:        question.Category,
		CategoryComment: categoryComment,
		Value:           question.Value,
		Type:            question.Type,
		Duration:        demoDuration,
	})
	return message.Message{
		Event:   domain.QuestionDemo,
		Payload: payload,
	}
}

func HandleQuestionDemoMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
