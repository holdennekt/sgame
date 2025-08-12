package message

import (
	"encoding/json"

	"github.com/holdennekt/sgame/internal/domain"
)

type Message struct {
	Id      string          `json:"id"`
	Event   domain.Event    `json:"event"`
	Payload json.RawMessage `json:"payload"`
}
