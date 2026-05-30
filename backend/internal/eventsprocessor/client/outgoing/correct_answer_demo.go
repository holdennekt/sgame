package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

func NewCorrectAnswerDemoMessage(toQuestionCorrectAnswerDemoer domain.ToQuestionCorrectAnswerDemoer, getAttachmentUrl func(key string) (string, error)) message.Message {
	questionCorrectAnswerDemo := toQuestionCorrectAnswerDemoer.ToQuestionCorrectAnswerDemo(getAttachmentUrl)
	payload, _ := json.Marshal(questionCorrectAnswerDemo)
	return message.Message{
		Event:   domain.CorrectAnswerDemo,
		Payload: payload,
	}
}

func HandleCorrectAnswerDemoMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
