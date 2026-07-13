package service

import (
	"context"
	"fmt"
	"log/slog"
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

// pendingURLUpload defers a URL fetch until its key's draft/pack is persisted, avoiding a notification racing ahead of registration.
type pendingURLUpload struct {
	key string
	url string
}

// A key already referenced elsewhere (e.g. pasted) keeps its metadata — no notification fires for it.
func (s *AttachmentService) createDomain(ctx context.Context, req dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, *pendingURLUpload, error) {
	key := req.Key
	var upload *pendingURLUpload
	if req.URL != "" {
		var err error
		key, err = s.generateKeyFromURL(req.URL, privacyType == domain.Public)
		if err != nil {
			return nil, nil, err
		}
		upload = &pendingURLUpload{key: key, url: req.URL}
	} else if existing := s.findRegistered(ctx, key); existing != nil {
		clone := *existing
		clone.URL = ""
		return &clone, nil, nil
	}
	return &domain.Attachment{
		Key:    key,
		Status: domain.AttachmentPending,
	}, upload, nil
}

// A failed fetch marks just that one attachment failed, not the whole (already-persisted) save.
func (s *AttachmentService) uploadPendingURLs(ctx context.Context, uploads []pendingURLUpload) {
	for _, u := range uploads {
		if err := s.rawStorage.UploadFromURL(ctx, storage.URLUploadInput{Key: u.key, URL: u.url, MaxBytes: MAX_FILE_SIZE}); err != nil {
			slog.Error("failed to fetch url attachment", "key", u.key, "url", u.url, "err", err)
			s.markFailed(ctx, u.key, "failed to fetch the provided URL")
		}
	}
}

// markFailed locates whichever draft/pack currently owns key and marks its attachment failed.
func (s *AttachmentService) markFailed(ctx context.Context, key, message string) {
	if draft, err := s.packDraftRepo.FindByAttachmentKey(ctx, key); err == nil {
		for attempt := 0; ; attempt++ {
			att := draft.GetAttachment(key)
			if att == nil {
				return
			}
			att.Status = domain.AttachmentFailed
			att.Error = message
			err := s.packDraftRepo.Update(ctx, draft)
			if err == nil {
				return
			}
			if _, ok := err.(custerr.ConflictErr); !ok || attempt >= SAVE_MAX_ATTEMPTS {
				slog.Error("failed to mark url attachment failed", "key", key, "err", err)
				return
			}
			draft, err = s.packDraftRepo.FindByAttachmentKey(ctx, key)
			if err != nil {
				slog.Error("failed to refetch draft to mark url attachment failed", "key", key, "err", err)
				return
			}
		}
	}

	pack, err := s.packRepo.FindByAttachmentKey(ctx, key)
	if err != nil {
		slog.Warn("could not locate draft/pack to mark url attachment failed", "key", key)
		return
	}
	att := pack.GetAttachment(key)
	if att == nil {
		return
	}
	att.Status = domain.AttachmentFailed
	att.Error = message
	if err := s.packRepo.Update(ctx, pack); err != nil {
		slog.Error("failed to mark url attachment failed", "key", key, "err", err)
	}
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

func (s *AttachmentService) upsertDomainDraft(ctx context.Context, oldAttachment *domain.Attachment, req dto.CreateAttachmentRequest) (*domain.Attachment, *pendingURLUpload, error) {
	if oldAttachment != nil {
		return oldAttachment, nil, nil
	}
	return s.createDomain(ctx, req, domain.Private)
}

func (s *AttachmentService) upsertDomain(ctx context.Context, oldAttachment *domain.Attachment, req dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, *pendingURLUpload, error) {
	if oldAttachment == nil {
		return s.createDomain(ctx, req, privacyType)
	}

	attachment := oldAttachment
	// only ready attachments have an object in the compressed bucket to re-key
	if oldAttachment.IsReady() {
		keyUUID := req.Key[strings.Index(req.Key, "/"):]
		newKey := string(privacyType) + keyUUID
		if oldAttachment.Key != newKey {
			if err := s.compressedStorage.Move(ctx, oldAttachment.Key, newKey); err != nil {
				return nil, nil, err
			}
			attachment.Key = newKey
		}
	}
	return attachment, nil, nil
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

// isAnimatedSource reports whether the probed content is a format that can
// itself carry multiple frames (gif, animated webp, apng). Everything else
// gets capped to its first frame during transcoding, even if ffmpeg's
// demuxer finds more than one — see transcodeImage.
func isAnimatedSource(data *ffprobe.ProbeData) bool {
	for _, stream := range data.Streams {
		if stream.CodecType != "video" {
			continue
		}
		switch stream.CodecName {
		case "gif", "webp", "apng":
			return true
		}
	}
	return false
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
