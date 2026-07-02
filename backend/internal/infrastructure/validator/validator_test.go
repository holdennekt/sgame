package validator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
)

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

func ollamaChatTestServer(t *testing.T, responses map[string]ivalidator.ValidatorResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(req.Messages) != 2 {
			http.Error(w, "expected system + user messages", http.StatusBadRequest)
			return
		}

		var payload struct {
			PlayerAnswer string `json:"player_answer"`
		}
		if err := json.Unmarshal([]byte(req.Messages[1].Content), &payload); err != nil {
			http.Error(w, "bad user content", http.StatusBadRequest)
			return
		}

		resp, ok := responses[payload.PlayerAnswer]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		content, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"content": string(content)},
		})
	}))
}

func TestLocalLLMValidator(t *testing.T) {
	tests := []struct {
		name         string
		playerAnswer string
		response     ivalidator.ValidatorResponse
		wantCorrect  bool
	}{
		{
			name:         "correct answer",
			playerAnswer: "Paris",
			response:     ivalidator.ValidatorResponse{IsCorrect: true, Confidence: 1.0},
			wantCorrect:  true,
		},
		{
			name:         "cross-script transliteration accepted",
			playerAnswer: "рхчп",
			response:     ivalidator.ValidatorResponse{IsCorrect: true, Confidence: 0.9},
			wantCorrect:  true,
		},
		{
			name:         "wrong answer",
			playerAnswer: "London",
			response:     ivalidator.ValidatorResponse{IsCorrect: false, Confidence: 0.95},
			wantCorrect:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := ollamaChatTestServer(t, map[string]ivalidator.ValidatorResponse{
				tt.playerAnswer: tt.response,
			})
			defer srv.Close()

			v := NewLocalLLMValidator(srv.URL, "test system prompt")
			q := ivalidator.Question{CorrectAnswers: []string{"Paris"}}
			resp, err := v.Validate(context.Background(), q, tt.playerAnswer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.IsCorrect != tt.wantCorrect {
				t.Errorf("IsCorrect = %v, want %v", resp.IsCorrect, tt.wantCorrect)
			}
		})
	}
}

func TestLocalLLMValidatorContextCancellation(t *testing.T) {
	srv := ollamaChatTestServer(t, map[string]ivalidator.ValidatorResponse{
		"Paris": {IsCorrect: true, Confidence: 1.0},
	})
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	q := ivalidator.Question{CorrectAnswers: []string{"Paris"}}
	v := NewLocalLLMValidator(srv.URL, "test system prompt")
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
