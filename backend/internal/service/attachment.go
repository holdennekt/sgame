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

func (s *AttachmentService) upsertDomainDraft(ctx context.Context, oldAttachment *domain.Attachment, req dto.CreateAttachmentRequest) (*domain.Attachment, error) {
	var attachment *domain.Attachment

	if oldAttachment != nil {
		attachment = oldAttachment
	} else {
		var err error
		attachment, err = s.createDomain(ctx, req, domain.Private)
		if err != nil {
			return nil, err
		}
	}
	return attachment, nil
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
		attachment, err = s.createDomain(ctx, req, privacyType)
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
	defer func() { _ = reader.Close() }()

	tmpFile, err := os.CreateTemp("", "sgame-probe-*")
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()
	if _, err := io.Copy(tmpFile, reader); err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to buffer file for probing: %w", err))
	}

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	probeData, err := ffprobe.ProbeURL(probeCtx, tmpFile.Name())
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to analyze media: %w", err))
	}

	fileType, mimeType := classifyProbe(probeData)
	att := &domain.Attachment{
		Key:      key,
		MimeType: mimeType,
		Size:     stats.Size,
		Type:     fileType,
	}
	if att.Type == domain.Video || att.Type == domain.Audio {
		att.Duration = probeData.Format.DurationSeconds
	} else {
		att.Duration = DEFAULT_ATTACHMENT_DURATION
	}

	if mimeType != stats.ContentType {
		if err := s.storage.SetContentType(ctx, key, mimeType); err != nil {
			return nil, err
		}
	}

	return att, nil
}

// classifyProbe derives both the coarse FileType and a concrete MIME type from
// ffprobe's analysis of the actual file content — more reliable than guessing
// from a filename extension, and works for misnamed or extensionless files.
func classifyProbe(data *ffprobe.ProbeData) (domain.FileType, string) {
	imageMIMETypes := map[string]string{
		"png": "image/png", "mjpeg": "image/jpeg", "jpeg": "image/jpeg",
		"gif": "image/gif", "bmp": "image/bmp", "webp": "image/webp",
		"tiff": "image/tiff", "svg": "image/svg+xml",
	}

	hasVideo, hasAudio := false, false
	imageMIME := "image/jpeg"
	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "video":
			if mimeType, ok := imageMIMETypes[stream.CodecName]; ok {
				imageMIME = mimeType
			} else {
				hasVideo = true
			}
		case "audio":
			hasAudio = true
		}
	}

	switch {
	case hasVideo:
		return domain.Video, containerMIMEType(data.Format.FormatName, domain.Video)
	case hasAudio:
		return domain.Audio, containerMIMEType(data.Format.FormatName, domain.Audio)
	default:
		return domain.Image, imageMIME
	}
}

// containerMIMEType maps ffprobe's (often comma-separated, ambiguous) container
// format name to a concrete MIME type, using fileType to disambiguate formats
// ffmpeg can't tell apart from the container alone (e.g. mp4 video vs m4a audio).
func containerMIMEType(formatName string, fileType domain.FileType) string {
	for name := range strings.SplitSeq(formatName, ",") {
		switch name {
		case "mp3":
			return "audio/mpeg"
		case "wav":
			return "audio/wav"
		case "ogg":
			return "audio/ogg"
		case "flac":
			return "audio/flac"
		case "webm", "matroska":
			if fileType == domain.Audio {
				return "audio/webm"
			}
			return "video/webm"
		case "mov", "mp4", "m4a", "3gp", "3g2", "mj2":
			if fileType == domain.Audio {
				return "audio/mp4"
			}
			return "video/mp4"
		}
	}
	if fileType == domain.Audio {
		return "audio/mpeg"
	}
	return "video/mp4"
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
