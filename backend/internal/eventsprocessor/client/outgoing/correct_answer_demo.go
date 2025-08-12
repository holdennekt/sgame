package outgoing

import (
	"context"
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

const CorrectAnswerDemoDuration = 5

type CorrectAnswerDemoPayload struct {
	Answers  []string `json:"answers"`
	Comment  *string  `json:"comment"`
	Duration int      `json:"duration"`
}

func NewCorrectAnswerDemoMessage(toQuestionCorrectAnswerer domain.ToQuestionCorrectAnswerer) message.Message {
	questionCorrectAnswer := toQuestionCorrectAnswerer.ToQuestionCorrectAnswer()
	payload, _ := json.Marshal(CorrectAnswerDemoPayload{
		Answers:  questionCorrectAnswer.Answers,
		Comment:  questionCorrectAnswer.Comment,
		Duration: CorrectAnswerDemoDuration,
	})
	return message.Message{
		Event:   domain.CorrectAnswerDemo,
		Payload: payload,
	}
}

func HandleCorrectAnswerDemoMessage(ctx context.Context, client realtime.Channel, msg message.Message) error {
	return client.Send(ctx, msg)
}
