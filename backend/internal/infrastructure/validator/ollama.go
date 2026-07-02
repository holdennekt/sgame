package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
)

const ollamaLLMModel = "llama3.2:3b"

type LocalLLMValidator struct {
	ollamaURL    string
	systemPrompt string
	client       *http.Client
}

func NewLocalLLMValidator(ollamaURL, systemPrompt string) *LocalLLMValidator {
	return &LocalLLMValidator{
		ollamaURL:    ollamaURL,
		systemPrompt: systemPrompt,
		client:       &http.Client{},
	}
}

func (v *LocalLLMValidator) Validate(ctx context.Context, question ivalidator.Question, playerAnswer string) (ivalidator.ValidatorResponse, error) {
	userContent, err := json.Marshal(map[string]any{
		"question_text":   question.Text,
		"correct_answers": question.CorrectAnswers,
		"player_answer":   playerAnswer,
	})
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	body, err := json.Marshal(map[string]any{
		"model":  ollamaLLMModel,
		"stream": false,
		"format": "json",
		"messages": []map[string]string{
			{"role": "system", "content": v.systemPrompt},
			{"role": "user", "content": string(userContent)},
		},
	})
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.ollamaURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ivalidator.ValidatorResponse{}, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	var validatorResp ivalidator.ValidatorResponse
	if err := json.Unmarshal([]byte(result.Message.Content), &validatorResp); err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	return validatorResp, nil
}
