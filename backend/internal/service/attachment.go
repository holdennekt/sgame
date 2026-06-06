package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type AttachmentService struct {
	storage storage.Storage
}

func NewAttachmentService(storage storage.Storage) *AttachmentService {
	return &AttachmentService{storage}
}

func (s *AttachmentService) createDomain(ctx context.Context, req dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, error) {
	key := req.Key
	if req.URL != "" {
		var err error
		key, err = s.generateKeyFromURL(req.URL, privacyType == domain.Public)
		if err != nil {
			return nil, err
		}
		if err = s.storage.UploadFromURL(ctx, storage.URLUploadInput{Key: key, URL: req.URL, MaxBytes: MAX_FILE_SIZE}); err != nil {
			return nil, err
		}
	}
	return s.probe(ctx, key)
}

func (s *AttachmentService) upsertDomain(ctx context.Context, oldAttachment *domain.Attachment, req dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, error) {
	var attachment *domain.Attachment

	if oldAttachment != nil {
		keyUUID := req.Key[strings.Index(req.Key, "/"):]
		newKey := string(privacyType) + keyUUID
		if oldAttachment.Key != newKey {
			if err := s.storage.Move(ctx, oldAttachment.Key, newKey); err != nil {
				return nil, err
			}
		}
		attachment = oldAttachment
		attachment.Key = newKey
	} else {
		var err error
		attachment, err = s.createDomain(
			ctx,
			req,
			privacyType,
		)
		if err != nil {
			return nil, err
		}
	}
	return attachment, nil
}

func (s *AttachmentService) probe(ctx context.Context, key string) (*domain.Attachment, error) {
	stats, err := s.storage.GetStats(ctx, key)
	if err != nil {
		return nil, err
	}
	reader, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tmpFile, err := os.CreateTemp("", "sgame-probe-*")
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	if _, err := io.Copy(tmpFile, reader); err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to buffer file for probing: %w", err))
	}

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	probeData, err := ffprobe.ProbeURL(probeCtx, tmpFile.Name())
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to analyze media: %w", err))
	}

	att := &domain.Attachment{
		Key:      key,
		MimeType: stats.ContentType,
		Size:     stats.Size,
		Type:     s.attachmentTypeFromProbe(probeData),
	}
	if att.Type == domain.Video || att.Type == domain.Audio {
		att.Duration = probeData.Format.DurationSeconds
	} else {
		att.Duration = DEFAULT_ATTACHMENT_DURATION
	}
	return att, nil
}

func (s *AttachmentService) attachmentTypeFromProbe(data *ffprobe.ProbeData) domain.FileType {
	imageCodecs := map[string]bool{
		"png": true, "mjpeg": true, "jpeg": true, "gif": true,
		"bmp": true, "webp": true, "tiff": true, "svg": true,
	}
	hasVideo, hasAudio := false, false
	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "video":
			if !imageCodecs[stream.CodecName] {
				hasVideo = true
			}
		case "audio":
			hasAudio = true
		}
	}
	if hasVideo {
		return domain.Video
	}
	if hasAudio {
		return domain.Audio
	}
	return domain.Image
}

func (s *AttachmentService) generateKey(filename string, public bool) string {
	ext := filepath.Ext(filename)
	id := uuid.New().String()
	prefix := "private"
	if public {
		prefix = "public"
	}

	return fmt.Sprintf("%s/%s%s", prefix, id, ext)
}

func (s *AttachmentService) generateKeyFromURL(sourceUrl string, public bool) (string, error) {
	u, err := url.Parse(sourceUrl)
	if err != nil {
		return "", custerr.NewBadRequestErr(fmt.Sprintf("error parsing URL: %v", err))
	}
	filename := filepath.Base(u.Path)
	return s.generateKey(filename, public), nil
}
