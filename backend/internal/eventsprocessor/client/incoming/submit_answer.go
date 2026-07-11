package incoming

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type SubmitAnswerPayload struct {
	Answer string `json:"answer"`
}

func HandleSubmitAnswerMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, roomId string, user domain.User, validator ivalidator.AnswerValidator, cfg *config.Config, msg message.Message) error {
	var sap SubmitAnswerPayload
	if err := json.Unmarshal(msg.Payload, &sap); err != nil {
		return err
	}

	var question domain.Question
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.CurrentQuestion != nil {
			question = room.CurrentQuestion.Question
		}
		return room.SubmitAnswer(user.Id, sap.Answer)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	go func() {
		iQuestion := buildValidatorQuestion(question, cfg.CompressedBucketName)
		isCorrect := resolveIsCorrect(validator, cfg, roomId, iQuestion, sap.Answer)

		finalCtx, finalCancel := context.WithTimeout(context.Background(), 5*time.Second)
		newRoom, err := roomCache.SafeUpdate(finalCtx, roomId, func(room *domain.Room) error {
			answeringEnded :=
				room.State != domain.Answering ||
					room.AnsweringPlayer.Id != user.Id ||
					!question.IsCurrent(room)
			if answeringEnded {
				return serverevent.ErrDeferredFunctionCancelled
			}
			if room.CurrentQuestion != nil {
				question = room.CurrentQuestion.Question
			}
			return room.ValidateAnswer(domain.SYSTEM, isCorrect)
		})
		finalCancel()

		if err != nil {
			if !errors.Is(err, serverevent.ErrDeferredFunctionCancelled) {
				slog.Error("error applying AI validation result", "err", err, "room_id", roomId)
			}
			return
		}

		if err := server.Send(context.Background(), outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
			slog.Error("error sending room_updated after AI validation", "err", err)
			return
		}

		if err := handlePostValidate(context.Background(), internalServer, newRoom, question); err != nil {
			slog.Error("error in post-validate handler", "err", err)
		}
	}()

	return nil
}

func isExactMatch(correctAnswers []string, playerAnswer string) bool {
	normalized := strings.ToLower(strings.TrimSpace(playerAnswer))
	for _, correct := range correctAnswers {
		if strings.ToLower(strings.TrimSpace(correct)) == normalized {
			return true
		}
	}
	return false
}

// resolveIsCorrect judges playerAnswer against iQuestion, checking for an
// exact match before falling back to the AI validator
func resolveIsCorrect(validator ivalidator.AnswerValidator, cfg *config.Config, roomId string, iQuestion ivalidator.Question, playerAnswer string) bool {
	if isExactMatch(iQuestion.CorrectAnswers, playerAnswer) {
		return true
	}

	timeout := time.Duration(cfg.AIValidationTimeout) * time.Second
	aiCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := validator.Validate(aiCtx, iQuestion, playerAnswer)
	if err != nil {
		slog.Error("AI validation failed, defaulting to wrong answer", "err", err, "room_id", roomId)
		return false
	}
	slog.Info("AI validation result",
		"room_id", roomId,
		"player_answer", playerAnswer,
		"correct_answers", iQuestion.CorrectAnswers,
		"is_correct", result.IsCorrect,
		"confidence", result.Confidence,
	)
	return result.IsCorrect
}

func buildValidatorQuestion(question domain.Question, bucketName string) ivalidator.Question {
	return buildValidatorQuestionFrom(question.Answers, question.Text, question.Attachment, bucketName)
}

func buildFinalRoundValidatorQuestion(question domain.FinalRoundQuestion, bucketName string) ivalidator.Question {
	return buildValidatorQuestionFrom(question.Answers, question.Text, question.Attachment, bucketName)
}

func buildValidatorQuestionFrom(answers []string, text *string, attachment *domain.Attachment, bucketName string) ivalidator.Question {
	q := ivalidator.Question{
		CorrectAnswers: answers,
	}
	if text != nil {
		q.Text = *text
	}
	if attachment != nil {
		q.MediaURI = "gs://" + bucketName + "/" + attachment.Key
		q.MediaMIMEType = attachment.MimeType
		switch attachment.Type {
		case domain.Image:
			q.MediaType = ivalidator.MediaTypeImage
		case domain.Audio:
			q.MediaType = ivalidator.MediaTypeAudio
		case domain.Video:
			q.MediaType = ivalidator.MediaTypeVideo
		}
	}
	return q
}
