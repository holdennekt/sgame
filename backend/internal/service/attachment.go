package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type AttachmentService struct {
	packDraftRepo     repository.PackDraft
	packRepo          repository.Pack
	rawStorage        storage.Raw
	compressedStorage storage.Compressed
}

func NewAttachmentService(packDraftRepo repository.PackDraft, packRepo repository.Pack, rawStorage storage.Raw, compressedStorage storage.Compressed) *AttachmentService {
	return &AttachmentService{packDraftRepo, packRepo, rawStorage, compressedStorage}
}

// A key already referenced elsewhere (e.g. pasted) keeps its metadata — no notification fires for it.
func (s *AttachmentService) createDomain(ctx context.Context, req dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, error) {
	key := req.Key
	if req.URL != "" {
		var err error
		key, err = s.generateKeyFromURL(req.URL, privacyType == domain.Public)
		if err != nil {
			return nil, err
		}
		if err = s.rawStorage.UploadFromURL(ctx, storage.URLUploadInput{Key: key, URL: req.URL, MaxBytes: MAX_FILE_SIZE}); err != nil {
			return nil, err
		}
	} else if existing := s.findRegistered(ctx, key); existing != nil {
		clone := *existing
		clone.URL = ""
		return &clone, nil
	}
	return &domain.Attachment{
		Key:    key,
		Status: domain.AttachmentPending,
	}, nil
}

func (s *AttachmentService) findRegistered(ctx context.Context, key string) *domain.Attachment {
	if draft, err := s.packDraftRepo.FindByAttachmentKey(ctx, key); err == nil {
		return draft.GetAttachment(key)
	}
	if pack, err := s.packRepo.FindByAttachmentKey(ctx, key); err == nil {
		return pack.GetAttachment(key)
	}
	return nil
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
		attachment = oldAttachment
		// only ready attachments have an object in the compressed bucket to re-key
		if oldAttachment.IsReady() {
			keyUUID := req.Key[strings.Index(req.Key, "/"):]
			newKey := string(privacyType) + keyUUID
			if oldAttachment.Key != newKey {
				if err := s.compressedStorage.Move(ctx, oldAttachment.Key, newKey); err != nil {
					return nil, err
				}
				attachment.Key = newKey
			}
		}
	} else {
		var err error
		attachment, err = s.createDomain(ctx, req, privacyType)
		if err != nil {
			return nil, err
		}
	}
	return attachment, nil
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

// No extension: compression normalizes the codec, so the source extension would mislead.
func (s *AttachmentService) generateKey(public bool) string {
	id := uuid.New().String()
	prefix := "private"
	if public {
		prefix = "public"
	}

	return fmt.Sprintf("%s/%s", prefix, id)
}

func (s *AttachmentService) generateKeyFromURL(sourceUrl string, public bool) (string, error) {
	if _, err := url.Parse(sourceUrl); err != nil {
		return "", custerr.NewBadRequestErr(fmt.Sprintf("error parsing URL: %v", err))
	}
	return s.generateKey(public), nil
}
