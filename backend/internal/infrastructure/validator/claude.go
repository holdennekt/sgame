package validator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	"github.com/anthropics/anthropic-sdk-go/lib/betafallback"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
	"golang.org/x/oauth2/google"
)

// Bedrock model IDs carry an "anthropic." prefix; Vertex ones are bare.
const (
	DefaultBedrockModel  = "anthropic.claude-haiku-5-5"
	DefaultVertexModel   = "claude-haiku-5-5"
	bedrockFallbackModel = "anthropic.claude-sonnet-5"
	vertexFallbackModel  = "claude-sonnet-5"
	claudeMaxTokens      = 4096
	// the Messages API rejects base64 images above 5 MB once encoded
	maxImageBytes = 3_750_000
)

var claudeImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type ClaudeValidator struct {
	messages     anthropic.BetaMessageService
	model        string
	systemPrompt string
	media        storage.Storage
}

func refusalFallback(model string) option.RequestOption {
	return option.WithMiddleware(betafallback.BetaRefusalFallbackMiddleware(
		[]anthropic.BetaFallbackParam{{Model: model}},
	))
}

func NewBedrockClaudeValidator(ctx context.Context, region, model, systemPrompt string, media storage.Storage) (*ClaudeValidator, error) {
	client, err := bedrock.NewMantleClient(ctx, bedrock.MantleClientConfig{AWSRegion: region}, refusalFallback(bedrockFallbackModel))
	if err != nil {
		return nil, err
	}
	if model == "" {
		model = DefaultBedrockModel
	}
	return &ClaudeValidator{messages: client.Beta.Messages, model: model, systemPrompt: systemPrompt, media: media}, nil
}

func NewVertexClaudeValidator(ctx context.Context, projectID, location, model, systemPrompt string, media storage.Storage) (*ClaudeValidator, error) {
	if location == "" {
		location = "global"
	}
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to find Google credentials: %w", err)
	}
	// the fallback middleware must run before the Vertex request rewrite
	client := anthropic.NewClient(refusalFallback(vertexFallbackModel), vertex.WithCredentials(ctx, location, projectID, creds))
	if model == "" {
		model = DefaultVertexModel
	}
	return &ClaudeValidator{messages: client.Beta.Messages, model: model, systemPrompt: systemPrompt, media: media}, nil
}

func (v *ClaudeValidator) Validate(ctx context.Context, question ivalidator.Question, playerAnswer string) (ivalidator.ValidatorResponse, error) {
	blocks, err := v.buildContent(ctx, question, playerAnswer)
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}

	resp, err := v.messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:        v.model,
		MaxTokens:    claudeMaxTokens,
		System:       []anthropic.BetaTextBlockParam{{Text: v.systemPrompt}},
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortLow},
		Messages:     []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(blocks...)},
	})
	if err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	if resp.StopReason == anthropic.BetaStopReasonRefusal {
		return ivalidator.ValidatorResponse{}, errors.New("model refused to validate the answer")
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(b.Text)
		}
	}
	return parseValidatorJSON(text.String())
}

// Only images are attached: Claude takes no audio, and video is skipped for latency.
func (v *ClaudeValidator) buildContent(ctx context.Context, question ivalidator.Question, playerAnswer string) ([]anthropic.BetaContentBlockParamUnion, error) {
	var blocks []anthropic.BetaContentBlockParamUnion

	if question.MediaKey != "" && question.MediaType == ivalidator.MediaTypeImage && claudeImageTypes[question.MediaMIMEType] {
		data, err := v.readImage(ctx, question.MediaKey)
		if err != nil {
			return nil, err
		}
		if data != nil {
			blocks = append(blocks, anthropic.NewBetaImageBlock(anthropic.BetaBase64ImageSourceParam{
				Data:      base64.StdEncoding.EncodeToString(data),
				MediaType: anthropic.BetaBase64ImageSourceMediaType(question.MediaMIMEType),
			}))
		}
	}

	payload, err := json.Marshal(map[string]any{
		"question_text":   question.Text,
		"correct_answers": question.CorrectAnswers,
		"player_answer":   playerAnswer,
	})
	if err != nil {
		return nil, err
	}
	blocks = append(blocks, anthropic.NewBetaTextBlock(string(payload)))
	return blocks, nil
}

// readImage returns nil for images too large to attach, so validation falls back to text only.
func (v *ClaudeValidator) readImage(ctx context.Context, key string) ([]byte, error) {
	r, err := v.media.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	data, err := io.ReadAll(io.LimitReader(r, maxImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read question image: %w", err)
	}
	if len(data) > maxImageBytes {
		return nil, nil
	}
	return data, nil
}

// Bedrock has no structured outputs, so tolerate prose or code fences around the object.
func parseValidatorJSON(text string) (ivalidator.ValidatorResponse, error) {
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return ivalidator.ValidatorResponse{}, fmt.Errorf("no JSON object in model response: %q", text)
	}
	var result ivalidator.ValidatorResponse
	if err := json.Unmarshal([]byte(text[start:end+1]), &result); err != nil {
		return ivalidator.ValidatorResponse{}, err
	}
	return result, nil
}
