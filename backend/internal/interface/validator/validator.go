package validator

import "context"

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeAudio MediaType = "audio"
	MediaTypeVideo MediaType = "video"
)

type Question struct {
	Text           string
	CorrectAnswers []string
	MediaURI       string // gs:// URI; empty for text-only questions
	MediaType      MediaType
}

type ValidatorResponse struct {
	IsCorrect  bool    `json:"is_correct"`
	Confidence float64 `json:"confidence"`
}

type AnswerValidator interface {
	Validate(ctx context.Context, question Question, playerAnswer string) (ValidatorResponse, error)
}
