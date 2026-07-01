package incoming

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"path"
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

	var capturedQuestion domain.Question
	var capturedPlayerId string
	_, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.CurrentQuestion != nil {
			capturedQuestion = room.CurrentQuestion.Question
		}
		if room.AnsweringPlayer != nil {
			capturedPlayerId = room.AnsweringPlayer.Id
		}
		return room.SubmitTypedAnswer(user.Id, sap.Answer)
	})
	if err != nil {
		return err
	}

	if err := server.Send(ctx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
		return err
	}

	go func() {
		timeout := time.Duration(cfg.AIValidationTimeout) * time.Second
		aiCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		iQuestion := buildValidatorQuestion(capturedQuestion, cfg.BucketName)

		isCorrect := false
		if isExactMatch(iQuestion.CorrectAnswers, sap.Answer) {
			isCorrect = true
		} else {
			result, err := validator.Validate(aiCtx, iQuestion, sap.Answer)
			if err != nil {
				slog.Error("AI validation failed, defaulting to wrong answer", "err", err, "room_id", roomId)
			} else {
				isCorrect = result.IsCorrect
			}
		}

		finalCtx, finalCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer finalCancel()

		var question domain.Question
		newRoom, err := roomCache.SafeUpdate(finalCtx, roomId, func(room *domain.Room) error {
			if room.State != domain.Answering || room.AnsweringPlayer == nil || room.AnsweringPlayer.Id != capturedPlayerId {
				return serverevent.ErrDeferredFunctionCancelled
			}
			if room.CurrentQuestion != nil {
				question = room.CurrentQuestion.Question
			}
			return room.ValidateAnswer(domain.SYSTEM, isCorrect)
		})
		if err != nil {
			if !errors.Is(err, serverevent.ErrDeferredFunctionCancelled) {
				slog.Error("error applying AI validation result", "err", err, "room_id", roomId)
			}
			return
		}

		if err := server.Send(finalCtx, outgoing.NewRoomUpdatedMessage(roomId)); err != nil {
			slog.Error("error sending room_updated after AI validation", "err", err)
			return
		}

		if err := handlePostValidate(finalCtx, internalServer, newRoom, question); err != nil {
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

func resolveMIMEType(stored, key string, attachType domain.FileType) string {
	if stored != "" {
		return stored
	}
	switch strings.ToLower(path.Ext(key)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	}
	switch attachType {
	case domain.Image:
		return "image/jpeg"
	case domain.Audio:
		return "audio/mpeg"
	case domain.Video:
		return "video/mp4"
	}
	return ""
}

func buildValidatorQuestion(question domain.Question, bucketName string) ivalidator.Question {
	q := ivalidator.Question{
		CorrectAnswers: question.Answers,
	}
	if question.Text != nil {
		q.Text = *question.Text
	}
	if question.Attachment != nil {
		q.MediaURI = "gs://" + bucketName + "/" + question.Attachment.Key
		q.MediaMIMEType = resolveMIMEType(question.Attachment.MimeType, question.Attachment.Key, question.Attachment.Type)
		switch question.Attachment.Type {
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
