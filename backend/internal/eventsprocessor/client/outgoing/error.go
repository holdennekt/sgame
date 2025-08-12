package outgoing

import (
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/message"
)

type errorPayload struct {
	Error string `json:"error"`
}

func NewErrorMessage(err error) message.Message {
	payload, _ := json.Marshal(errorPayload{Error: err.Error()})
	return message.Message{
		Event:   domain.Error,
		Payload: payload,
	}
}
