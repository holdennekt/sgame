package service

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/holdennekt/sgame/backend/pkg/sets"
	"gopkg.in/vansante/go-ffprobe.v2"
)

const (
	GET_URL_TTL                 = 3 * time.Hour
	POST_URL_TTL                = 5 * time.Minute
	MAX_FILE_SIZE               = 200 << 20 // 200 MB
	DEFAULT_ATTACHMENT_DURATION = 1
)

type PackService struct {
	packRepository repository.Pack
	storage        storage.Storage
}

func NewPackService(packRepository repository.Pack, storage storage.Storage) *PackService {
	return &PackService{packRepository, storage}
}

func (s *PackService) Create(ctx context.Context, user domain.User, cpr dto.CreatePackRequest) (string, error) {
	if user.IsGuest {
		return "", custerr.NewForbiddenErr("guest users aren't allowed to create packs")
	}

	var err error

	defer func() {
		if err != nil {
			for key := range cpr.AttachmentKeys() {
				if err := s.storage.Delete(context.Background(), key); err != nil {
					log.Printf("failed to cleanup attachment %s: %v", key, err)
				}
			}
		}
	}()

	pack, err := s.createDomain(ctx, cpr, user)
	if err != nil {
		return "", err
	}

	return s.packRepository.Create(ctx, pack)
}

func (s *PackService) GetById(ctx context.Context, userId, id string) (*domain.Pack, error) {
	pack, err := s.packRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if pack.Type != "public" && userId != pack.CreatedBy.Id && userId != domain.SYSTEM {
		return nil, custerr.NewForbiddenErr("can not get private pack")
	}

	for ri, round := range pack.Rounds {
		for ci, category := range round.Categories {
			for qi := range category.Questions {
				att := pack.Rounds[ri].Categories[ci].Questions[qi].Attachment
				if att == nil {
					continue
				}
				u, err := s.storage.URL(ctx, att.Key, GET_URL_TTL)
				if err != nil {
					return nil, err
				}
				pack.Rounds[ri].Categories[ci].Questions[qi].Attachment.URL = u
			}
		}
	}
	for ci := range pack.FinalRound.Categories {
		att := pack.FinalRound.Categories[ci].Question.Attachment
		if att == nil {
			continue
		}
		u, err := s.storage.URL(ctx, att.Key, GET_URL_TTL)
		if err != nil {
			return nil, err
		}
		pack.FinalRound.Categories[ci].Question.Attachment.URL = u
	}

	return pack, nil
}

func (s *PackService) GetPreviews(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackPreview, int, error) {
	return s.packRepository.GetPreviews(ctx, userId, search)
}

func (s *PackService) GetHiddens(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.HiddenPack, int, error) {
	return s.packRepository.GetHiddens(ctx, userId, search)
}

func (s *PackService) GetCreatedBy(ctx context.Context, userId, createdBy string, search dto.SearchRequest) ([]domain.HiddenPack, int, error) {
	return s.packRepository.GetCreatedBy(ctx, userId, createdBy, search)
}

func (s *PackService) Update(ctx context.Context, user domain.User, dto dto.UpdatePackRequest) error {
	pack, err := s.packRepository.GetById(ctx, dto.Id)
	if err != nil {
		return err
	}

	if pack.CreatedBy.Id != user.Id {
		return custerr.NewForbiddenErr("can only edit your own packs")
	}

	defer func() {
		if err != nil {
			for key := range sets.Delta(dto.CreatePackRequest.AttachmentKeys(), pack.AttachmentKeys()) {
				if err := s.storage.Delete(context.Background(), key); err != nil {
					log.Printf("failed to cleanup attachment %s: %v", key, err)
				}
			}
		}
	}()

	newPack, err := s.updateDomain(ctx, pack, dto, user)
	if err != nil {
		return err
	}

	err = s.packRepository.Update(ctx, newPack)
	if err != nil {
		return err
	}

	for key := range sets.Delta(pack.AttachmentKeys(), dto.CreatePackRequest.AttachmentKeys()) {
		if err := s.storage.Delete(context.Background(), key); err != nil {
			log.Printf("failed to cleanup attachment %s: %v", key, err)
		}
	}
	return nil
}

func (s *PackService) Delete(ctx context.Context, userId, id string) error {
	pack, err := s.packRepository.GetById(ctx, id)
	if err != nil {
		return err
	}

	if pack.CreatedBy.Id != userId {
		return custerr.NewForbiddenErr("can only delete your own packs")
	}

	for key := range pack.AttachmentKeys() {
		if err := s.storage.Delete(context.Background(), key); err != nil {
			log.Printf("failed to cleanup attachment %s: %v", key, err)
		}
	}

	return s.packRepository.Delete(ctx, id)
}

func (s *PackService) SignURL(ctx context.Context, user domain.User, req dto.SignURLRequest) (*storage.SignUploadPolicyResult, string, error) {
	if user.IsGuest {
		return nil, "", custerr.NewForbiddenErr("guest users aren't allowed to upload media")
	}

	key := s.generateKey(req.Filename, *req.Public)
	result, err := s.storage.SignUploadPolicy(ctx, storage.SignUploadPolicyInput{Key: key, TTL: POST_URL_TTL})
	if err != nil {
		return nil, "", err
	}
	getUrl := ""
	if *req.Public {
		getUrl, err = s.storage.URL(ctx, key, GET_URL_TTL)
		if err != nil {
			return nil, "", err
		}
	}
	return result, getUrl, nil
}

func (s *PackService) validateRoundsCheckSum(ctx context.Context, userId string, pack dto.CreatePackRequest, ignoreId string) ([]byte, error) {
	marshaledRounds, _ := json.Marshal(struct {
		Rounds     []dto.CreateRoundRequest    `json:"rounds"`
		FinalRound dto.CreateFinalRoundRequest `json:"finalRound"`
	}{Rounds: pack.Rounds, FinalRound: pack.FinalRound})
	hasher := sha256.New()
	_, err := hasher.Write(marshaledRounds)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	roundsChecksum, err := hasher.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}

	packs, err := s.packRepository.GetByChecksum(
		ctx,
		userId,
		roundsChecksum,
		ignoreId,
	)
	if err != nil {
		return nil, err
	}
	if len(packs) == 0 {
		return roundsChecksum, nil
	}
	return nil, custerr.NewConflictErr(
		fmt.Sprintf("the pack with such rounds already exists and has id \"%s\"", packs[0].Id),
	)
}

func (s *PackService) createDomain(ctx context.Context, dto dto.CreatePackRequest, user domain.User) (*domain.Pack, error) {
	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, dto, "")
	if err != nil {
		return nil, err
	}

	content := []string{dto.Name}
	rounds := []domain.Round{}
	for _, r := range dto.Rounds {
		content = append(content, r.Name)
		round := domain.Round{
			Name:       r.Name,
			Categories: []domain.Category{},
		}
		for _, c := range r.Categories {
			content = append(content, c.Name)
			category := domain.Category{
				Name:      c.Name,
				Questions: []domain.Question{},
			}
			for _, q := range c.Questions {
				question := domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:      r.Name,
						Category:   c.Name,
						Index:      q.Index,
						Value:      q.Value,
						Attachment: nil,
					},
					Type:    q.Type,
					Text:    q.Text,
					Answers: q.Answers,
				}

				if q.Attachment != nil {
					attachment, err := s.createDomainAttachment(ctx, *q.Attachment, dto.Type)
					if err != nil {
						return nil, err
					}
					question.Attachment = attachment
				}

				if q.Comment != nil {
					question.Comment = &domain.Comment{
						Text: q.Comment.Text,
					}
					if q.Comment.Attachment != nil {
						attachment, err := s.createDomainAttachment(ctx, *q.Comment.Attachment, dto.Type)
						if err != nil {
							return nil, err
						}
						question.Comment.Attachment = attachment
					}
				}

				category.Questions = append(category.Questions, question)
			}
			round.Categories = append(round.Categories, category)
		}
		rounds = append(rounds, round)
	}

	finalRound := domain.FinalRound{
		Categories: []domain.FinalRoundCategory{},
	}
	for _, c := range dto.FinalRound.Categories {
		category := domain.FinalRoundCategory{
			HiddenFinalRoundCategory: domain.HiddenFinalRoundCategory{
				Name: c.Name,
			},
			Question: domain.FinalRoundQuestion{
				HiddenFinalRoundQuestion: domain.HiddenFinalRoundQuestion{
					Category: c.Name,
					Text:     c.Question.Text,
				},
				Answers: c.Question.Answers,
			},
		}

		if c.Question.Attachment != nil {
			attachment, err := s.createDomainAttachment(ctx, *c.Question.Attachment, dto.Type)
			if err != nil {
				return nil, err
			}
			category.Question.Attachment = attachment
		}

		if c.Question.Comment != nil {
			category.Question.Comment = &domain.Comment{
				Text: c.Question.Comment.Text,
			}
			if c.Question.Comment.Attachment != nil {
				attachment, err := s.createDomainAttachment(ctx, *c.Question.Comment.Attachment, dto.Type)
				if err != nil {
					return nil, err
				}
				category.Question.Comment.Attachment = attachment
			}
		}

		finalRound.Categories = append(finalRound.Categories, category)
	}

	return &domain.Pack{
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        strings.Join(content, ", "),
		Name:           dto.Name,
		Type:           dto.Type,
		Rounds:         rounds,
		FinalRound:     finalRound,
	}, nil
}

func (s *PackService) updateDomain(ctx context.Context, oldPack *domain.Pack, dto dto.UpdatePackRequest, user domain.User) (*domain.Pack, error) {
	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, dto.CreatePackRequest, dto.Id)
	if err != nil {
		return nil, err
	}

	content := []string{dto.Name}
	rounds := []domain.Round{}
	for _, r := range dto.Rounds {
		content = append(content, r.Name)
		round := domain.Round{
			Name:       r.Name,
			Categories: []domain.Category{},
		}
		for _, c := range r.Categories {
			content = append(content, c.Name)
			category := domain.Category{
				Name:      c.Name,
				Questions: []domain.Question{},
			}
			for _, q := range c.Questions {
				question := domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:      r.Name,
						Category:   c.Name,
						Index:      q.Index,
						Value:      q.Value,
						Attachment: nil,
					},
					Type:    q.Type,
					Text:    q.Text,
					Answers: q.Answers,
				}

				if q.Attachment != nil {
					attachment, err := s.upsertDomainAttachment(ctx, oldPack, *q.Attachment, dto.Type)
					if err != nil {
						return nil, err
					}
					question.Attachment = attachment
				}

				if q.Comment != nil {
					question.Comment = &domain.Comment{
						Text: q.Comment.Text,
					}
					if q.Comment.Attachment != nil {
						attachment, err := s.upsertDomainAttachment(ctx, oldPack, *q.Comment.Attachment, dto.Type)
						if err != nil {
							return nil, err
						}
						question.Comment.Attachment = attachment
					}
				}

				category.Questions = append(category.Questions, question)
			}
			round.Categories = append(round.Categories, category)
		}
		rounds = append(rounds, round)
	}

	finalRound := domain.FinalRound{
		Categories: []domain.FinalRoundCategory{},
	}
	for _, c := range dto.FinalRound.Categories {
		category := domain.FinalRoundCategory{
			HiddenFinalRoundCategory: domain.HiddenFinalRoundCategory{
				Name: c.Name,
			},
			Question: domain.FinalRoundQuestion{
				HiddenFinalRoundQuestion: domain.HiddenFinalRoundQuestion{
					Category: c.Name,
					Text:     c.Question.Text,
				},
				Answers: c.Question.Answers,
			},
		}

		if c.Question.Attachment != nil {
			attachment, err := s.upsertDomainAttachment(ctx, oldPack, *c.Question.Attachment, dto.Type)
			if err != nil {
				return nil, err
			}
			category.Question.Attachment = attachment
		}

		if c.Question.Comment != nil {
			category.Question.Comment = &domain.Comment{
				Text: c.Question.Comment.Text,
			}
			if c.Question.Comment.Attachment != nil {
				attachment, err := s.upsertDomainAttachment(ctx, oldPack, *c.Question.Comment.Attachment, dto.Type)
				if err != nil {
					return nil, err
				}
				category.Question.Comment.Attachment = attachment
			}
		}

		finalRound.Categories = append(finalRound.Categories, category)
	}

	return &domain.Pack{
		Id:             dto.Id,
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        strings.Join(content, ", "),
		Name:           dto.Name,
		Type:           dto.Type,
		Rounds:         rounds,
		FinalRound:     finalRound,
	}, nil
}

func (s *PackService) createDomainAttachment(ctx context.Context, dto dto.CreateAttachmentRequest, privacyType domain.PrivacyType) (*domain.Attachment, error) {
	attachment := &domain.Attachment{
		Key: dto.Key,
	}
	if dto.URL != "" {
		key, err := s.generateKeyFromURL(dto.URL, privacyType == domain.Public)
		if err != nil {
			return nil, err
		}
		err = s.storage.UploadFromURL(
			ctx,
			storage.URLUploadInput{
				Key:      key,
				URL:      dto.URL,
				MaxBytes: MAX_FILE_SIZE,
			},
		)
		if err != nil {
			return nil, err
		}

		attachment.Key = key
	}

	attachmentStats, err := s.storage.GetStats(ctx, attachment.Key)
	if err != nil {
		return nil, err
	}
	attachment.MimeType = attachmentStats.ContentType
	attachment.Size = attachmentStats.Size

	reader, err := s.storage.Get(ctx, attachment.Key)
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

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	probeData, err := ffprobe.ProbeURL(ctx, tmpFile.Name())
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to analyze media: %w", err))
	}
	attachment.Type = s.getAttachmentType(probeData)

	if attachment.Type == domain.Video || attachment.Type == domain.Audio {
		attachment.Duration = probeData.Format.DurationSeconds
	} else {
		attachment.Duration = DEFAULT_ATTACHMENT_DURATION
	}

	return attachment, nil
}

func (s *PackService) upsertDomainAttachment(ctx context.Context, oldPack *domain.Pack, newAttachmentRequest dto.CreateAttachmentRequest, newPrivacyType domain.PrivacyType) (*domain.Attachment, error) {
	var attachment *domain.Attachment

	oldAttachment := oldPack.GetAttachment(newAttachmentRequest.Key)
	if oldAttachment != nil {
		keyUUID := newAttachmentRequest.Key[strings.Index(newAttachmentRequest.Key, "/"):]
		newKey := string(newPrivacyType) + keyUUID
		if oldAttachment.Key != newKey {
			if err := s.storage.Move(ctx, oldAttachment.Key, newKey); err != nil {
				return nil, err
			}
		}
		attachment = oldAttachment
		attachment.Key = newKey
	} else {
		var err error
		attachment, err = s.createDomainAttachment(
			ctx,
			newAttachmentRequest,
			newPrivacyType,
		)
		if err != nil {
			return nil, err
		}
	}
	return attachment, nil
}

func (s *PackService) getAttachmentType(data *ffprobe.ProbeData) domain.FileType {
	hasVideo := false
	hasAudio := false

	imageCodecs := map[string]bool{
		"png":   true,
		"mjpeg": true,
		"jpeg":  true,
		"gif":   true,
		"bmp":   true,
		"webp":  true,
		"tiff":  true,
		"svg":   true,
	}

	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "video":
			if _, ok := imageCodecs[stream.CodecName]; !ok {
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

func (s *PackService) generateKey(filename string, public bool) string {
	ext := filepath.Ext(filename)
	id := uuid.New().String()
	prefix := "private"
	if public {
		prefix = "public"
	}

	return fmt.Sprintf("%s/%s%s", prefix, id, ext)
}

func (s *PackService) generateKeyFromURL(sourceUrl string, public bool) (string, error) {
	u, err := url.Parse(sourceUrl)
	if err != nil {
		return "", custerr.NewBadRequestErr(fmt.Sprintf("error parsing URL: %v", err))
	}
	filename := filepath.Base(u.Path)
	return s.generateKey(filename, public), nil
}
