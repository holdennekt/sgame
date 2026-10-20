package validator

import (
	"context"
	"encoding/json"

	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
	"google.golang.org/genai"
)

const geminiModel = "gemini-3.5-flash"

type GeminiValidator struct {
	client            *genai.Client
	systemInstruction *genai.Content
	mediaBucket       string
}

func NewGeminiValidator(ctx context.Context, projectID, location, systemInstruction, mediaBucket string) (*GeminiValidator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, err
	}
	return &GeminiValidator{
		client: client,
		systemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemInstruction}},
		},
		mediaBucket: mediaBucket,
	}, nil
}

func (v *GeminiValidator) Validate(ctx context.Context, question ivalidator.Question, playerAnswer string) (ivalidator.ValidatorResponse, error) {
	parts, err := buildPromptParts(question, v.mediaBucket, playerAnswer)
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	resp, err := v.client.Models.GenerateContent(
		ctx,
		geminiModel,
		[]*genai.Content{{Parts: parts, Role: "user"}},
		&genai.GenerateContentConfig{
			SystemInstruction: v.systemInstruction,
			ResponseMIMEType:  "application/json",
		},
	)
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	var result ivalidator.ValidatorResponse
	if err := json.Unmarshal([]byte(resp.Text()), &result); err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	return result, nil
}

// buildPromptParts assembles the multimodal content parts for the model.
// Video questions skip the media entirely — processing latency is too high for
// a live game; the correct answers carry enough signal on their own.
func buildPromptParts(question ivalidator.Question, mediaBucket, playerAnswer string) ([]*genai.Part, error) {
	var parts []*genai.Part

	if question.MediaKey != "" && question.MediaType != ivalidator.MediaTypeVideo {
		parts = append(parts, &genai.Part{
			FileData: &genai.FileData{
				FileURI:  "gs://" + mediaBucket + "/" + question.MediaKey,
				MIMEType: question.MediaMIMEType,
			},
		})
	}

	payload, err := json.Marshal(map[string]any{
		"question_text":   question.Text,
		"correct_answers": question.CorrectAnswers,
		"player_answer":   playerAnswer,
	})
	if err != nil {
		return nil, err
	}
	parts = append(parts, &genai.Part{Text: string(payload)})

	return parts, nil
}
