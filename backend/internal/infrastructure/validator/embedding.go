package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
)

const ollamaEmbeddingModel = "mxbai-embed-large"

type LocalEmbeddingValidator struct {
	ollamaURL string
	threshold float64
	client    *http.Client
}

func NewLocalEmbeddingValidator(ollamaURL string, threshold float64) *LocalEmbeddingValidator {
	return &LocalEmbeddingValidator{
		ollamaURL: ollamaURL,
		threshold: threshold,
		client:    &http.Client{},
	}
}

func (v *LocalEmbeddingValidator) Validate(ctx context.Context, question ivalidator.Question, playerAnswer string) (ivalidator.ValidatorResponse, error) {
	var best ivalidator.ValidatorResponse
	for _, correctAnswer := range question.CorrectAnswers {
		r, err := v.validatePair(ctx, correctAnswer, playerAnswer)
		if err != nil {
			return ivalidator.ValidatorResponse{}, err
		}
		fmt.Println(r)
		if r.Confidence > best.Confidence {
			best = r
		}
	}
	return best, nil
}

func (v *LocalEmbeddingValidator) validatePair(ctx context.Context, correctAnswer, playerAnswer string) (ivalidator.ValidatorResponse, error) {
	correctEmb, err := v.embed(ctx, correctAnswer)
	if err != nil {
		return ivalidator.ValidatorResponse{}, fmt.Errorf("embedding correct answer %q: %w", correctAnswer, err)
	}
	playerEmb, err := v.embed(ctx, playerAnswer)
	if err != nil {
		return ivalidator.ValidatorResponse{}, fmt.Errorf("embedding player answer: %w", err)
	}

	similarity, ok := cosineSimilarity(correctEmb, playerEmb)
	if !ok {
		return ivalidator.ValidatorResponse{}, nil
	}
	return ivalidator.ValidatorResponse{
		IsCorrect:  similarity >= v.threshold,
		Confidence: similarity,
	}, nil
}

func (v *LocalEmbeddingValidator) embed(ctx context.Context, text string) ([]float64, error) {
	body, err := json.Marshal(map[string]string{
		"model":  ollamaEmbeddingModel,
		"prompt": text,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.ollamaURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}

func cosineSimilarity(a, b []float64) (float64, bool) {
	if len(a) == 0 || len(a) != len(b) {
		return 0, false
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0, false
	}
	return dot / denom, true
}
