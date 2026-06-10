package service

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/holdennekt/sgame/backend/pkg/sets"
)

const (
	GET_URL_TTL                 = 3 * time.Hour
	POST_URL_TTL                = 5 * time.Minute
	MAX_FILE_SIZE               = 200 << 20 // 200 MB
	DEFAULT_ATTACHMENT_DURATION = 1
)

type PackService struct {
	packRepository    repository.Pack
	storage           storage.Storage
	attachmentService *AttachmentService
}

func NewPackService(packRepository repository.Pack, storage storage.Storage, attachmentService *AttachmentService) *PackService {
	return &PackService{packRepository, storage, attachmentService}
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
					slog.Error("failed to cleanup attachment", "key", key, "err", err)
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

func (s *PackService) CreateFromDraft(ctx context.Context, user domain.User, draft *domain.PackDraft) (string, error) {
	if user.IsGuest {
		return "", custerr.NewForbiddenErr("guest users aren't allowed to create packs")
	}

	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, draftToCreatePackRequest(draft), nil)
	if err != nil {
		return "", err
	}

	now := time.Now()
	return s.packRepository.Create(ctx, &domain.Pack{
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        draft.Content,
		Name:           draft.Name,
		Type:           draft.Type,
		Rounds:         draft.Rounds,
		FinalRound:     draft.FinalRound,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func (s *PackService) GetById(ctx context.Context, userId, id string) (any, error) {
	pack, err := s.packRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if userId != pack.CreatedBy.Id && userId != domain.SYSTEM {
		if pack.Type != domain.Public {
			return nil, custerr.NewForbiddenErr("can not get private pack")
		}
		return domain.NewHiddenPack(*pack), nil
	}

	if err := populatePackAttachmentURLs(ctx, s.storage, pack); err != nil {
		return nil, err
	}
	return pack, nil
}

func populatePackAttachmentURLs(ctx context.Context, stor storage.Storage, pack *domain.Pack) error {
	setURL := func(a *domain.Attachment) error {
		if a == nil {
			return nil
		}
		u, err := stor.URL(ctx, a.Key, GET_URL_TTL)
		if err != nil {
			return err
		}
		a.URL = u
		return nil
	}
	for ri := range pack.Rounds {
		for ci := range pack.Rounds[ri].Categories {
			for qi := range pack.Rounds[ri].Categories[ci].Questions {
				q := &pack.Rounds[ri].Categories[ci].Questions[qi]
				if err := setURL(q.Attachment); err != nil {
					return err
				}
				if q.Comment != nil {
					if err := setURL(q.Comment.Attachment); err != nil {
						return err
					}
				}
			}
		}
	}
	for ci := range pack.FinalRound.Categories {
		q := &pack.FinalRound.Categories[ci].Question
		if err := setURL(q.Attachment); err != nil {
			return err
		}
		if q.Comment != nil {
			if err := setURL(q.Comment.Attachment); err != nil {
				return err
			}
		}
	}
	return nil
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

func (s *PackService) UpdateFromDraft(ctx context.Context, user domain.User, draft *domain.PackDraft) error {
	oldPack, err := s.packRepository.GetById(ctx, *draft.LinkedPackId)
	if err != nil {
		return err
	}
	if oldPack.CreatedBy.Id != user.Id {
		return custerr.NewForbiddenErr("can only edit your own packs")
	}

	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, draftToCreatePackRequest(draft), draft.LinkedPackId)
	if err != nil {
		return err
	}

	if err := s.packRepository.Update(ctx, &domain.Pack{
		Id:             oldPack.Id,
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        draft.Content,
		Name:           draft.Name,
		Type:           draft.Type,
		Rounds:         draft.Rounds,
		FinalRound:     draft.FinalRound,
		CreatedAt:      oldPack.CreatedAt,
		UpdatedAt:      time.Now(),
	}); err != nil {
		return err
	}

	for key := range sets.Delta(oldPack.AttachmentKeys(), draft.AttachmentKeys()) {
		if err := s.storage.Delete(context.Background(), key); err != nil {
			slog.Error("failed to cleanup attachment", "key", key, "err", err)
		}
	}
	return nil
}

func (s *PackService) Update(ctx context.Context, user domain.User, req dto.UpdatePackRequest) error {
	pack, err := s.packRepository.GetById(ctx, req.Id)
	if err != nil {
		return err
	}

	if pack.CreatedBy.Id != user.Id {
		return custerr.NewForbiddenErr("can only edit your own packs")
	}

	defer func() {
		if err != nil {
			for key := range sets.Delta(req.AttachmentKeys(), pack.AttachmentKeys()) {
				if err := s.storage.Delete(context.Background(), key); err != nil {
					slog.Error("failed to cleanup attachment", "key", key, "err", err)
				}
			}
		}
	}()

	newPack, err := s.updateDomain(ctx, user, pack, req)
	if err != nil {
		return err
	}

	err = s.packRepository.Update(ctx, newPack)
	if err != nil {
		return err
	}

	for key := range sets.Delta(pack.AttachmentKeys(), req.AttachmentKeys()) {
		if err := s.storage.Delete(context.Background(), key); err != nil {
			slog.Error("failed to cleanup attachment", "key", key, "err", err)
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

	if err := s.packRepository.Delete(ctx, id); err != nil {
		return err
	}

	go func() {
		for key := range pack.AttachmentKeys() {
			if err := s.storage.Delete(context.Background(), key); err != nil {
				slog.Error("failed to cleanup attachment", "key", key, "err", err)
			}
		}
	}()

	return nil
}

func (s *PackService) SignURL(ctx context.Context, user domain.User, req dto.SignURLRequest) (*storage.SignUploadPolicyResult, string, error) {
	if user.IsGuest {
		return nil, "", custerr.NewForbiddenErr("guest users aren't allowed to upload media")
	}

	key := s.attachmentService.generateKey(req.Filename, *req.Public)
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

func (s *PackService) validateRoundsCheckSum(ctx context.Context, userId string, pack dto.CreatePackRequest, ignoreId *string) ([]byte, error) {
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
	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, dto, nil)
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
				Comment:   c.Comment,
				Questions: []domain.Question{},
			}
			for _, q := range c.Questions {
				question := domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:    r.Name,
						Category: c.Name,
						Index:    q.Index,
						Value:    q.Value,
					},
					Type:       q.Type,
					Text:       q.Text,
					Attachment: nil,
					Answers:    q.Answers,
				}

				if q.Attachment != nil {
					attachment, err := s.attachmentService.createDomain(ctx, *q.Attachment, dto.Type)
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
						attachment, err := s.attachmentService.createDomain(ctx, *q.Comment.Attachment, dto.Type)
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
			attachment, err := s.attachmentService.createDomain(ctx, *c.Question.Attachment, dto.Type)
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
				attachment, err := s.attachmentService.createDomain(ctx, *c.Question.Comment.Attachment, dto.Type)
				if err != nil {
					return nil, err
				}
				category.Question.Comment.Attachment = attachment
			}
		}

		finalRound.Categories = append(finalRound.Categories, category)
	}

	now := time.Now()
	return &domain.Pack{
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        strings.Join(content, ", "),
		Name:           dto.Name,
		Type:           dto.Type,
		Rounds:         rounds,
		FinalRound:     finalRound,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (s *PackService) updateDomain(ctx context.Context, user domain.User, oldPack *domain.Pack, req dto.UpdatePackRequest) (*domain.Pack, error) {
	roundsChecksum, err := s.validateRoundsCheckSum(ctx, user.Id, req.CreatePackRequest, &req.Id)
	if err != nil {
		return nil, err
	}

	content := []string{req.Name}
	rounds := []domain.Round{}
	for _, r := range req.Rounds {
		content = append(content, r.Name)
		round := domain.Round{
			Name:       r.Name,
			Categories: []domain.Category{},
		}
		for _, c := range r.Categories {
			content = append(content, c.Name)
			category := domain.Category{
				Name:      c.Name,
				Comment:   c.Comment,
				Questions: []domain.Question{},
			}
			for _, q := range c.Questions {
				question := domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:    r.Name,
						Category: c.Name,
						Index:    q.Index,
						Value:    q.Value,
					},
					Type:       q.Type,
					Text:       q.Text,
					Attachment: nil,
					Answers:    q.Answers,
				}

				if q.Attachment != nil {
					attachment, err := s.attachmentService.upsertDomain(ctx, oldPack.GetAttachment(q.Attachment.Key), *q.Attachment, req.Type)
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
						attachment, err := s.attachmentService.upsertDomain(ctx, oldPack.GetAttachment(q.Comment.Attachment.Key), *q.Comment.Attachment, req.Type)
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
	for _, c := range req.FinalRound.Categories {
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
			attachment, err := s.attachmentService.upsertDomain(ctx, oldPack.GetAttachment(c.Question.Attachment.Key), *c.Question.Attachment, req.Type)
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
				attachment, err := s.attachmentService.upsertDomain(ctx, oldPack.GetAttachment(c.Question.Comment.Attachment.Key), *c.Question.Comment.Attachment, req.Type)
				if err != nil {
					return nil, err
				}
				category.Question.Comment.Attachment = attachment
			}
		}

		finalRound.Categories = append(finalRound.Categories, category)
	}

	return &domain.Pack{
		Id:             req.Id,
		CreatedBy:      user,
		RoundsChecksum: roundsChecksum,
		Content:        strings.Join(content, ", "),
		Name:           req.Name,
		Type:           req.Type,
		Rounds:         rounds,
		FinalRound:     finalRound,
		CreatedAt:      oldPack.CreatedAt,
		UpdatedAt:      time.Now(),
	}, nil
}
