package validator

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		a, b      []float64
		wantScore float64
		wantOK    bool
	}{
		{
			name:      "identical vectors",
			a:         []float64{1, 0, 0},
			b:         []float64{1, 0, 0},
			wantScore: 1.0,
			wantOK:    true,
		},
		{
			name:      "orthogonal vectors",
			a:         []float64{1, 0},
			b:         []float64{0, 1},
			wantScore: 0.0,
			wantOK:    true,
		},
		{
			name:      "opposite vectors",
			a:         []float64{1, 0},
			b:         []float64{-1, 0},
			wantScore: -1.0,
			wantOK:    true,
		},
		{
			name:      "similar non-unit vectors",
			a:         []float64{3, 4},
			b:         []float64{4, 3},
			wantScore: 24.0 / 25.0,
			wantOK:    true,
		},
		{
			name:      "scaled identical direction",
			a:         []float64{2, 0},
			b:         []float64{5, 0},
			wantScore: 1.0,
			wantOK:    true,
		},
		{
			name:   "zero vector a",
			a:      []float64{0, 0},
			b:      []float64{1, 0},
			wantOK: false,
		},
		{
			name:   "zero vector b",
			a:      []float64{1, 0},
			b:      []float64{0, 0},
			wantOK: false,
		},
		{
			name:   "both zero vectors",
			a:      []float64{0, 0},
			b:      []float64{0, 0},
			wantOK: false,
		},
		{
			name:   "empty vectors",
			a:      []float64{},
			b:      []float64{},
			wantOK: false,
		},
		{
			name:   "mismatched lengths",
			a:      []float64{1, 2, 3},
			b:      []float64{1, 2},
			wantOK: false,
		},
		{
			name:      "negative components",
			a:         []float64{-1, 0},
			b:         []float64{-1, 0},
			wantScore: 1.0,
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := cosineSimilarity(tt.a, tt.b)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && math.Abs(got-tt.wantScore) > 1e-9 {
				t.Errorf("score = %v, want %v", got, tt.wantScore)
			}
		})
	}
}

func TestBuildPromptParts(t *testing.T) {
	tests := []struct {
		name          string
		question      ivalidator.Question
		wantPartCount int
		wantFileData  bool
	}{
		{
			name: "text-only question produces one text part",
			question: ivalidator.Question{
				Text:           "What is the capital of France?",
				CorrectAnswers: []string{"Paris"},
			},
			wantPartCount: 1,
			wantFileData:  false,
		},
		{
			name: "image question prepends a file part",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Eiffel Tower"},
				MediaURI:       "gs://bucket/img.jpg",
				MediaType:      ivalidator.MediaTypeImage,
			},
			wantPartCount: 2,
			wantFileData:  true,
		},
		{
			name: "audio question prepends a file part",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Beethoven"},
				MediaURI:       "gs://bucket/clip.mp3",
				MediaType:      ivalidator.MediaTypeAudio,
			},
			wantPartCount: 2,
			wantFileData:  true,
		},
		{
			name: "video question skips media due to latency",
			question: ivalidator.Question{
				CorrectAnswers: []string{"answer"},
				MediaURI:       "gs://bucket/vid.mp4",
				MediaType:      ivalidator.MediaTypeVideo,
			},
			wantPartCount: 1,
			wantFileData:  false,
		},
		{
			name: "text + image both included",
			question: ivalidator.Question{
				Text:           "Name this landmark",
				CorrectAnswers: []string{"Eiffel Tower", "Tour Eiffel"},
				MediaURI:       "gs://bucket/tower.jpg",
				MediaType:      ivalidator.MediaTypeImage,
			},
			wantPartCount: 2,
			wantFileData:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := buildPromptParts(tt.question, "player answer")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(parts) != tt.wantPartCount {
				t.Errorf("part count = %d, want %d", len(parts), tt.wantPartCount)
			}
			hasFileData := len(parts) > 0 && parts[0].FileData != nil
			if hasFileData != tt.wantFileData {
				t.Errorf("hasFileData = %v, want %v", hasFileData, tt.wantFileData)
			}
			last := parts[len(parts)-1]
			if last.Text == "" {
				t.Error("last part should be a non-empty JSON text payload")
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(last.Text), &payload); err != nil {
				t.Errorf("last part is not valid JSON: %v", err)
			}
			if _, ok := payload["correct_answers"]; !ok {
				t.Error("payload missing correct_answers field")
			}
		})
	}
}

func ollamaTestServer(t *testing.T, embeddings map[string][]float64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		emb, ok := embeddings[req.Prompt]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"embedding": emb})
	}))
}

func TestLocalEmbeddingValidator(t *testing.T) {
	const threshold = 0.8

	tests := []struct {
		name         string
		question     ivalidator.Question
		playerAnswer string
		embeddings   map[string][]float64
		wantCorrect  bool
		wantErr      bool
	}{
		{
			name: "exact match returns similarity 1.0",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris"},
			},
			playerAnswer: "Paris",
			embeddings: map[string][]float64{
				"Paris": {1, 0, 0},
			},
			wantCorrect: true,
		},
		{
			name: "similar answer above threshold",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris"},
			},
			playerAnswer: "paris city",
			embeddings: map[string][]float64{
				"Paris":      {3, 4},
				"paris city": {4, 3},
			},
			// similarity = 24/25 = 0.96 >= 0.8
			wantCorrect: true,
		},
		{
			name: "dissimilar answer below threshold",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris"},
			},
			playerAnswer: "London",
			embeddings: map[string][]float64{
				"Paris":  {1, 0},
				"London": {0, 1},
			},
			// similarity = 0.0 < 0.8
			wantCorrect: false,
		},
		{
			name: "matches second correct answer",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris", "City of Light"},
			},
			playerAnswer: "City of Light",
			embeddings: map[string][]float64{
				"Paris":         {1, 0},
				"City of Light": {0, 1},
			},
			// player answer is identical to second correct answer → similarity 1.0
			wantCorrect: true,
		},
		{
			name: "best confidence across multiple correct answers is returned",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris", "City of Light"},
			},
			playerAnswer: "paris city",
			embeddings: map[string][]float64{
				"Paris":         {3, 4},  // similarity with "paris city" = 0.96
				"City of Light": {0, 1},  // similarity with "paris city" = 0.6
				"paris city":    {4, 3},
			},
			wantCorrect: true,
		},
		{
			name: "unknown correct answer returns error",
			question: ivalidator.Question{
				CorrectAnswers: []string{"unknown"},
			},
			playerAnswer: "anything",
			embeddings:   map[string][]float64{},
			wantErr:      true,
		},
		{
			name: "unknown player answer returns error",
			question: ivalidator.Question{
				CorrectAnswers: []string{"Paris"},
			},
			playerAnswer: "unknown",
			embeddings: map[string][]float64{
				"Paris": {1, 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := ollamaTestServer(t, tt.embeddings)
			defer srv.Close()

			v := NewLocalEmbeddingValidator(srv.URL, threshold)
			resp, err := v.Validate(context.Background(), tt.question, tt.playerAnswer)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.IsCorrect != tt.wantCorrect {
				t.Errorf("IsCorrect = %v, want %v (confidence = %.4f)", resp.IsCorrect, tt.wantCorrect, resp.Confidence)
			}
			if resp.Confidence < 0 || resp.Confidence > 1 {
				t.Errorf("confidence %v out of [0,1] range", resp.Confidence)
			}
		})
	}
}

func TestLocalEmbeddingValidatorContextCancellation(t *testing.T) {
	srv := ollamaTestServer(t, map[string][]float64{
		"Paris": {1, 0},
	})
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	q := ivalidator.Question{CorrectAnswers: []string{"Paris"}}
	v := NewLocalEmbeddingValidator(srv.URL, 0.8)
	_, err := v.Validate(ctx, q, "Paris")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

// MockValidator satisfies AnswerValidator for use in higher-level tests.
type MockValidator struct {
	Response ivalidator.ValidatorResponse
	Err      error
}

func (m *MockValidator) Validate(_ context.Context, _ ivalidator.Question, _ string) (ivalidator.ValidatorResponse, error) {
	return m.Response, m.Err
}

var _ ivalidator.AnswerValidator = (*MockValidator)(nil)
