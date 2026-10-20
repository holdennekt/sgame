package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/holdennekt/sgame/backend/pkg/metrics"
	"gopkg.in/vansante/go-ffprobe.v2"
)

const (
	NOT_REGISTERED_MAX_ATTEMPTS = 8
	NOT_REGISTERED_ATTEMPTS_TTL = 2 * time.Hour
	SAVE_MAX_ATTEMPTS           = 5
	TRANSCODE_TIMEOUT           = time.Minute
)

func DraftAttachmentsChannel(draftId string) string {
	return fmt.Sprintf("draft:%s:attachments", draftId)
}

// AttachmentProcessorService handles at-least-once upload notifications, so every step must stay idempotent.
type AttachmentProcessorService struct {
	packDraftRepo     repository.PackDraft
	packRepo          repository.Pack
	rawStorage        storage.Raw
	compressedStorage storage.Compressed
	channels          realtime.ChannelGetter
	attempts          cache.AttemptCounter
}

func NewAttachmentProcessorService(
	packDraftRepo repository.PackDraft,
	packRepo repository.Pack,
	rawStorage storage.Raw,
	compressedStorage storage.Compressed,
	channels realtime.ChannelGetter,
	attempts cache.AttemptCounter,
) *AttachmentProcessorService {
	return &AttachmentProcessorService{packDraftRepo, packRepo, rawStorage, compressedStorage, channels, attempts}
}

// A nil return settles the notification; an error asks for redelivery (outcome "retry" in Prometheus).
func (s *AttachmentProcessorService) Process(ctx context.Context, key string) (err error) {
	start := time.Now()
	slog.Info("processing attachment", "key", key)
	metrics.AttachmentsProcessingInFlight.Inc()
	outcome := "ready"
	defer func() {
		metrics.AttachmentsProcessingInFlight.Dec()
		duration := time.Since(start)
		metrics.AttachmentProcessingDuration.Observe(duration.Seconds())
		if err != nil {
			outcome = "retry"
		}
		metrics.AttachmentsProcessedTotal.WithLabelValues(outcome).Inc()
		slog.Info("finished processing attachment", "key", key, "outcome", outcome, "duration", duration.Round(time.Millisecond).String())
	}()

	doc, err := s.locate(ctx, key)
	if err != nil {
		if _, ok := err.(custerr.NotFoundErr); !ok {
			return err
		}
		outcome = "dropped"
		return s.handleNotRegistered(ctx, key)
	}

	att := doc.getAttachment(key)
	if att.Status != domain.AttachmentPending {
		outcome = "skipped"
		return nil
	}

	result, contentErr, err := s.processMedia(ctx, s.rawStorage, key)
	if err != nil {
		return err
	}

	if contentErr != nil {
		slog.Error("attachment processing failed", "key", key, "err", contentErr)
		att.Status = domain.AttachmentFailed
		outcome = "failed"
		if errors.Is(contentErr, errSourceGone) {
			att.Error = "the upload never completed or has expired, try uploading again"
		} else {
			att.Error = "unsupported or corrupt media file, try a different one"
		}
	} else {
		att.Type = result.fileType
		att.MimeType = result.mimeType
		att.Size = result.size
		att.Duration = result.duration
		att.Status = domain.AttachmentReady
		att.Error = ""
	}

	if err := s.save(ctx, doc, key, *att); err != nil {
		return err
	}

	// nothing will ever read the raw copy again, on failure included
	if err := s.rawStorage.Delete(ctx, key); err != nil {
		if _, ok := err.(custerr.NotFoundErr); !ok {
			slog.Error("failed to delete raw attachment, lifecycle TTL will collect it", "key", key, "err", err)
		}
	}

	s.publish(ctx, doc, *att)
	return nil
}

type attachmentDoc struct {
	draft *domain.PackDraft
	pack  *domain.Pack
}

func (d attachmentDoc) getAttachment(key string) *domain.Attachment {
	if d.draft != nil {
		return d.draft.GetAttachment(key)
	}
	return d.pack.GetAttachment(key)
}

func (s *AttachmentProcessorService) locate(ctx context.Context, key string) (attachmentDoc, error) {
	draft, err := s.packDraftRepo.FindByAttachmentKey(ctx, key)
	if err == nil {
		return attachmentDoc{draft: draft}, nil
	}
	if _, ok := err.(custerr.NotFoundErr); !ok {
		return attachmentDoc{}, err
	}
	pack, err := s.packRepo.FindByAttachmentKey(ctx, key)
	if err != nil {
		return attachmentDoc{}, err
	}
	return attachmentDoc{pack: pack}, nil
}

func (s *AttachmentProcessorService) handleNotRegistered(ctx context.Context, key string) error {
	attempts, err := s.attempts.Increment(ctx, "attachment:not-registered:"+key, NOT_REGISTERED_ATTEMPTS_TTL)
	if err != nil {
		slog.Error("failed to count not-registered attempts", "key", key, "err", err)
	}
	if err == nil && attempts > NOT_REGISTERED_MAX_ATTEMPTS {
		slog.Warn("dropping notification for unregistered attachment key; raw-bucket TTL will collect the object", "key", key, "attempts", attempts)
		return nil
	}
	return custerr.NewNotFoundErr(fmt.Sprintf("attachment key %q is not registered yet", key))
}

var errSourceGone = errors.New("source object is gone")

type processedMedia struct {
	fileType domain.FileType
	mimeType string
	size     int64
	duration float64
}

// processMedia downloads, probes, transcodes and uploads; contentErr is terminal, err is transient.
func (s *AttachmentProcessorService) processMedia(ctx context.Context, src storage.Storage, key string) (result processedMedia, contentErr error, err error) {
	inPath, err := s.download(ctx, src, key)
	if err != nil {
		if _, ok := err.(custerr.NotFoundErr); ok {
			return result, fmt.Errorf("%w: %v", errSourceGone, err), nil
		}
		return result, nil, err
	}
	defer func() { _ = os.Remove(inPath) }()

	probeData, probeErr := probeFile(ctx, inPath)
	if probeErr != nil {
		return result, fmt.Errorf("failed to analyze media: %w", probeErr), nil
	}

	outFile, err := os.CreateTemp("", "sgame-compressed-*")
	if err != nil {
		return result, nil, custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	_ = outFile.Close()
	outPath := outFile.Name()
	defer func() { _ = os.Remove(outPath) }()

	result, transcodeErr := s.transcode(ctx, probeData, inPath, outPath)
	if transcodeErr != nil {
		var ce ContentError
		if !errors.As(transcodeErr, &ce) {
			return result, nil, custerr.NewInternalErr(transcodeErr)
		}
		return result, transcodeErr, nil
	}

	result.size, err = s.uploadCompressed(ctx, key, result, outPath)
	if err != nil {
		return result, nil, err
	}
	return result, nil, nil
}

func (s *AttachmentProcessorService) download(ctx context.Context, src storage.Storage, key string) (string, error) {
	reader, err := src.Get(ctx, key)
	if err != nil {
		return "", err
	}
	defer func() { _ = reader.Close() }()

	inFile, err := os.CreateTemp("", "sgame-source-*")
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer func() { _ = inFile.Close() }()
	if _, err := io.Copy(inFile, reader); err != nil {
		_ = os.Remove(inFile.Name())
		return "", custerr.NewInternalErr(fmt.Errorf("failed to buffer source object: %w", err))
	}
	return inFile.Name(), nil
}

func probeFile(ctx context.Context, path string) (*ffprobe.ProbeData, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return ffprobe.ProbeURL(probeCtx, path)
}

func (s *AttachmentProcessorService) transcode(ctx context.Context, probeData *ffprobe.ProbeData, inPath, outPath string) (processedMedia, error) {
	fileType, _ := classifyProbe(probeData)
	sourceBitRate, _ := strconv.Atoi(probeData.Format.BitRate)

	transcodeCtx, cancel := context.WithTimeout(ctx, TRANSCODE_TIMEOUT)
	defer cancel()

	result := processedMedia{fileType: fileType, duration: DEFAULT_ATTACHMENT_DURATION}
	var err error
	switch fileType {
	case domain.Image:
		result.mimeType = IMAGE_MIME_TYPE
		err = transcodeImage(transcodeCtx, inPath, outPath, isAnimatedSource(probeData))
	case domain.Audio:
		result.mimeType = AUDIO_MIME_TYPE
		result.duration = probeData.Format.DurationSeconds
		err = transcodeAudio(transcodeCtx, inPath, outPath, sourceBitRate)
	case domain.Video:
		result.mimeType = VIDEO_MIME_TYPE
		result.duration = probeData.Format.DurationSeconds
		err = transcodeVideo(transcodeCtx, inPath, outPath, sourceBitRate)
	}
	return result, err
}

func (s *AttachmentProcessorService) uploadCompressed(ctx context.Context, key string, media processedMedia, path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, custerr.NewInternalErr(fmt.Errorf("failed to stat compressed file: %w", err))
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, custerr.NewInternalErr(fmt.Errorf("failed to open compressed file: %w", err))
	}
	defer func() { _ = f.Close() }()

	if err := s.compressedStorage.Upload(ctx, storage.UploadInput{
		Key:      key,
		Size:     stat.Size(),
		MimeType: media.mimeType,
		Type:     media.fileType,
		Reader:   f,
	}); err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// save retries version conflicts against concurrent draft autosaves.
func (s *AttachmentProcessorService) save(ctx context.Context, doc attachmentDoc, key string, att domain.Attachment) error {
	if doc.pack != nil {
		target := doc.pack.GetAttachment(key)
		if target == nil {
			slog.Warn("attachment key no longer present on pack, skipping save", "key", key)
			return nil
		}
		*target = att
		return s.packRepo.Update(ctx, doc.pack)
	}

	draft := doc.draft
	for attempt := 0; ; attempt++ {
		target := draft.GetAttachment(key)
		if target == nil {
			// a concurrent edit removed or replaced this attachment since we located it
			slog.Warn("attachment key no longer present on draft, skipping save", "key", key)
			return nil
		}
		*target = att
		err := s.packDraftRepo.Update(ctx, draft)
		if err == nil {
			doc.draft = draft
			return nil
		}
		if _, ok := err.(custerr.ConflictErr); !ok || attempt >= SAVE_MAX_ATTEMPTS {
			return err
		}
		draft, err = s.packDraftRepo.FindByAttachmentKey(ctx, key)
		if err != nil {
			if _, ok := err.(custerr.NotFoundErr); ok {
				slog.Warn("attachment key disappeared while saving processing result", "key", key)
				return nil
			}
			return err
		}
	}
}

func (s *AttachmentProcessorService) publish(ctx context.Context, doc attachmentDoc, att domain.Attachment) {
	if doc.draft == nil {
		return
	}
	if att.Status == domain.AttachmentReady {
		u, err := s.compressedStorage.URL(ctx, att.Key, GET_URL_TTL)
		if err != nil {
			slog.Error("failed to get URL for attachment", "key", att.Key, "err", err)
		} else {
			att.URL = u
		}
	}
	payload, _ := json.Marshal(att)
	channel := s.channels.Get(DraftAttachmentsChannel(doc.draft.Id))
	if err := channel.Send(ctx, message.Message{
		Id:      uuid.NewString(),
		Event:   domain.AttachmentUpdated,
		Payload: payload,
	}); err != nil {
		slog.Error("failed to publish attachment update", "key", att.Key, "draft_id", doc.draft.Id, "err", err)
	}
}
