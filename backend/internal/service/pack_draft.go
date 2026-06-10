package service

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/holdennekt/sgame/backend/pkg/custvalid"
	"github.com/holdennekt/sgame/backend/pkg/sets"
)

type PackDraftService struct {
	packDraftRepo     repository.PackDraft
	packRepo          repository.Pack
	storage           storage.Storage
	attachmentService *AttachmentService
	packService       *PackService
	validator         *validator.Validate
}

func NewPackDraftService(packDraftRepo repository.PackDraft, packRepo repository.Pack, storage storage.Storage, attachmentService *AttachmentService, packService *PackService) *PackDraftService {
	v := validator.New()
	v.SetTagName("binding")
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	v.RegisterValidation(custvalid.SameLength, custvalid.ValidateSameLength)
	return &PackDraftService{
		packDraftRepo:     packDraftRepo,
		packRepo:          packRepo,
		storage:           storage,
		attachmentService: attachmentService,
		packService:       packService,
		validator:         v,
	}
}

func (s *PackDraftService) GetOrCreateEditDraft(ctx context.Context, user domain.User, packId string) (string, error) {
	if user.IsGuest {
		return "", custerr.NewForbiddenErr("guest users cannot edit packs")
	}

	var pack *domain.Pack
	var err error
	if packId != "" {
		pack, err = s.packRepo.GetById(ctx, packId)
		if err != nil {
			return "", err
		}
		if pack.CreatedBy.Id != user.Id {
			return "", custerr.NewForbiddenErr("can only edit your own packs")
		}
	}

	existing, err := s.packDraftRepo.GetByUserAndLinkedPack(ctx, user.Id, packId)
	if err == nil {
		return existing.Id, nil
	}
	if _, ok := err.(custerr.NotFoundErr); !ok {
		return "", err
	}

	now := time.Now()
	if packId == "" {
		return s.packDraftRepo.Create(ctx, &domain.PackDraft{
			CreatedBy:  user,
			Name:       "New Pack",
			Type:       domain.Public,
			Rounds:     []domain.Round{{Name: "Round 1", Categories: []domain.Category{}}},
			FinalRound: domain.FinalRound{Categories: []domain.FinalRoundCategory{}},
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	return s.packDraftRepo.Create(ctx, &domain.PackDraft{
		LinkedPackId: &pack.Id,
		CreatedBy:    user,
		Name:         pack.Name,
		Type:         pack.Type,
		Rounds:       pack.Rounds,
		FinalRound:   pack.FinalRound,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *PackDraftService) GetById(ctx context.Context, userId, id string) (*domain.PackDraft, error) {
	draft, err := s.packDraftRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if draft.CreatedBy.Id != userId {
		return nil, custerr.NewForbiddenErr("cannot access another user's draft")
	}
	populateAttachmentURLs(ctx, s.storage, draft)
	return draft, nil
}

func populateAttachmentURLs(ctx context.Context, stor storage.Storage, draft *domain.PackDraft) {
	setURL := func(a *domain.Attachment) {
		if a == nil {
			return
		}
		u, err := stor.URL(ctx, a.Key, GET_URL_TTL)
		if err != nil {
			slog.Error("failed to get URL for attachment", "key", a.Key, "err", err)
			return
		}
		a.URL = u
	}
	for ri := range draft.Rounds {
		for ci := range draft.Rounds[ri].Categories {
			for qi := range draft.Rounds[ri].Categories[ci].Questions {
				q := &draft.Rounds[ri].Categories[ci].Questions[qi]
				setURL(q.Attachment)
				if q.Comment != nil {
					setURL(q.Comment.Attachment)
				}
			}
		}
	}
	for ci := range draft.FinalRound.Categories {
		q := &draft.FinalRound.Categories[ci].Question
		setURL(q.Attachment)
		if q.Comment != nil {
			setURL(q.Comment.Attachment)
		}
	}
}

func (s *PackDraftService) GetByUser(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackDraft, int, error) {
	return s.packDraftRepo.GetByUser(ctx, userId, search)
}

func (s *PackDraftService) Update(ctx context.Context, user domain.User, id string, req dto.UpdatePackDraftRequest) (*domain.PackDraft, error) {
	draft, err := s.packDraftRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if draft.CreatedBy.Id != user.Id {
		return nil, custerr.NewForbiddenErr("cannot edit another user's draft")
	}

	var linkedPackKeys map[string]struct{}
	skipCleanup := false
	if draft.LinkedPackId != nil {
		linkedPack, fetchErr := s.packRepo.GetById(context.Background(), *draft.LinkedPackId)
		if fetchErr != nil {
			slog.Error("draft update: failed to fetch linked pack, skipping attachment cleanup", "pack_id", *draft.LinkedPackId, "err", fetchErr)
			skipCleanup = true
		} else {
			linkedPackKeys = linkedPack.AttachmentKeys()
		}
	}

	defer func() {
		if err != nil && !skipCleanup {
			for key := range sets.Delta(sets.Delta(req.AttachmentKeys(), draft.AttachmentKeys()), linkedPackKeys) {
				if err := s.storage.Delete(context.Background(), key); err != nil {
					slog.Error("failed to cleanup attachment", "key", key, "err", err)
				}
			}
		}
	}()

	newDraft, err := s.updateDomain(ctx, user, draft, req)
	if err != nil {
		return nil, err
	}

	err = s.packDraftRepo.Update(ctx, newDraft)
	if err != nil {
		return nil, err
	}

	if !skipCleanup {
		for key := range sets.Delta(sets.Delta(draft.AttachmentKeys(), req.AttachmentKeys()), linkedPackKeys) {
			if err := s.storage.Delete(context.Background(), key); err != nil {
				slog.Error("failed to cleanup attachment", "key", key, "err", err)
			}
		}
	}

	populateAttachmentURLs(ctx, s.storage, newDraft)
	return newDraft, nil
}

func (s *PackDraftService) Delete(ctx context.Context, userId, id string) error {
	draft, err := s.packDraftRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if draft.CreatedBy.Id != userId {
		return custerr.NewForbiddenErr("cannot delete another user's draft")
	}

	if err := s.packDraftRepo.Delete(ctx, id); err != nil {
		return err
	}

	go func() {
		var linkedPackKeys map[string]struct{}
		if draft.LinkedPackId != nil {
			linkedPack, err := s.packRepo.GetById(context.Background(), *draft.LinkedPackId)
			if err != nil {
				slog.Error("draft delete: failed to fetch linked pack, skipping attachment cleanup", "pack_id", *draft.LinkedPackId, "err", err)
				return
			}
			linkedPackKeys = linkedPack.AttachmentKeys()
		}
		for key := range sets.Delta(draft.AttachmentKeys(), linkedPackKeys) {
			if err := s.storage.Delete(context.Background(), key); err != nil {
				slog.Error("failed to cleanup attachment", "key", key, "err", err)
			}
		}
	}()

	return nil
}

func (s *PackDraftService) Publish(ctx context.Context, user domain.User, id string) (string, error) {
	draft, err := s.packDraftRepo.GetById(ctx, id)
	if err != nil {
		return "", err
	}
	if draft.CreatedBy.Id != user.Id {
		return "", custerr.NewForbiddenErr("cannot publish another user's draft")
	}

	createPackRequest := draftToCreatePackRequest(draft)

	if err := s.validator.Struct(createPackRequest); err != nil {
		return "", err
	}

	if err := s.promoteAttachments(ctx, draft); err != nil {
		return "", err
	}
	if err := s.packDraftRepo.Update(ctx, draft); err != nil {
		return "", err
	}

	var packId string

	if draft.LinkedPackId != nil {
		if err := s.packService.UpdateFromDraft(ctx, user, draft); err != nil {
			return "", err
		}
		packId = *draft.LinkedPackId
	} else {
		var err error
		packId, err = s.packService.CreateFromDraft(ctx, user, draft)
		if err != nil {
			return "", err
		}
	}

	if err := s.packDraftRepo.Delete(ctx, id); err != nil {
		slog.Error("publish: failed to delete draft", "draft_id", id, "err", err)
	}
	return packId, nil
}

func (s *PackDraftService) promoteAttachments(ctx context.Context, draft *domain.PackDraft) error {
	targetType := draft.Type

	var atts []*domain.Attachment
	for ri := range draft.Rounds {
		for ci := range draft.Rounds[ri].Categories {
			for qi := range draft.Rounds[ri].Categories[ci].Questions {
				q := &draft.Rounds[ri].Categories[ci].Questions[qi]
				if q.Attachment != nil {
					atts = append(atts, q.Attachment)
				}
				if q.Comment != nil && q.Comment.Attachment != nil {
					atts = append(atts, q.Comment.Attachment)
				}
			}
		}
	}
	for ci := range draft.FinalRound.Categories {
		q := &draft.FinalRound.Categories[ci].Question
		if q.Attachment != nil {
			atts = append(atts, q.Attachment)
		}
		if q.Comment != nil && q.Comment.Attachment != nil {
			atts = append(atts, q.Comment.Attachment)
		}
	}

	if len(atts) == 0 {
		return nil
	}

	type result struct {
		att    *domain.Attachment
		newKey string
		err    error
	}

	jobs := make(chan *domain.Attachment)
	results := make(chan result)

	for range min(8, len(atts)) {
		go func() {
			for att := range jobs {
				slashIdx := strings.Index(att.Key, "/")
				newKey := string(targetType) + att.Key[slashIdx:]
				if att.Key == newKey {
					results <- result{att: att, newKey: newKey}
					continue
				}
				results <- result{att: att, newKey: newKey, err: s.storage.Move(ctx, att.Key, newKey)}
			}
		}()
	}

	go func() {
		for _, att := range atts {
			jobs <- att
		}
		close(jobs)
	}()

	var firstErr error
	for range atts {
		r := <-results
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
		} else {
			r.att.Key = r.newKey
		}
	}
	return firstErr
}

func (s *PackDraftService) updateDomain(ctx context.Context, user domain.User, oldDraft *domain.PackDraft, req dto.UpdatePackDraftRequest) (*domain.PackDraft, error) {
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
					attachment, err := s.attachmentService.upsertDomainDraft(ctx, oldDraft.GetAttachment(q.Attachment.Key), *q.Attachment)
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
						attachment, err := s.attachmentService.upsertDomainDraft(ctx, oldDraft.GetAttachment(q.Comment.Attachment.Key), *q.Comment.Attachment)
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
			attachment, err := s.attachmentService.upsertDomainDraft(ctx, oldDraft.GetAttachment(c.Question.Attachment.Key), *c.Question.Attachment)
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
				attachment, err := s.attachmentService.upsertDomainDraft(ctx, oldDraft.GetAttachment(c.Question.Comment.Attachment.Key), *c.Question.Comment.Attachment)
				if err != nil {
					return nil, err
				}
				category.Question.Comment.Attachment = attachment
			}
		}

		finalRound.Categories = append(finalRound.Categories, category)
	}

	return &domain.PackDraft{
		Id:           oldDraft.Id,
		LinkedPackId: oldDraft.LinkedPackId,
		CreatedBy:    user,
		Content:      strings.Join(content, ", "),
		Name:         req.Name,
		Type:         req.Type,
		Rounds:       rounds,
		FinalRound:   finalRound,
		CreatedAt:    oldDraft.CreatedAt,
		UpdatedAt:    time.Now(),
	}, nil
}

func draftToCreatePackRequest(draft *domain.PackDraft) dto.CreatePackRequest {
	rounds := make([]dto.CreateRoundRequest, len(draft.Rounds))
	for ri, r := range draft.Rounds {
		cats := make([]dto.CreateCategoryRequest, len(r.Categories))
		for ci, c := range r.Categories {
			qs := make([]dto.CreateQuestionRequest, len(c.Questions))
			for qi, q := range c.Questions {
				qs[qi] = dto.CreateQuestionRequest{
					Index:      q.Index,
					Value:      q.Value,
					Type:       q.Type,
					Text:       q.Text,
					Attachment: attToDTO(q.Attachment),
					Answers:    q.Answers,
					Comment:    commentToDTO(q.Comment),
				}
			}
			cats[ci] = dto.CreateCategoryRequest{Name: c.Name, Comment: c.Comment, Questions: qs}
		}
		rounds[ri] = dto.CreateRoundRequest{Name: r.Name, Categories: cats}
	}

	finalCats := make([]dto.CreateFinalRoundCategoryRequest, len(draft.FinalRound.Categories))
	for i, c := range draft.FinalRound.Categories {
		finalCats[i] = dto.CreateFinalRoundCategoryRequest{
			Name: c.Name,
			Question: dto.CreateFinalRoundQuestionRequest{
				Text:       c.Question.Text,
				Attachment: attToDTO(c.Question.Attachment),
				Answers:    c.Question.Answers,
				Comment:    commentToDTO(c.Question.Comment),
			},
		}
	}

	return dto.CreatePackRequest{
		Name:       draft.Name,
		Type:       draft.Type,
		Rounds:     rounds,
		FinalRound: dto.CreateFinalRoundRequest{Categories: finalCats},
	}
}

func attToDTO(a *domain.Attachment) *dto.CreateAttachmentRequest {
	if a == nil {
		return nil
	}
	return &dto.CreateAttachmentRequest{Key: a.Key}
}

func commentToDTO(c *domain.Comment) *dto.CreateCommentRequest {
	if c == nil {
		return nil
	}
	return &dto.CreateCommentRequest{Text: c.Text, Attachment: attToDTO(c.Attachment)}
}

func (s *PackDraftService) Import(ctx context.Context, user domain.User, r io.ReaderAt, size int64) (string, error) {
	if user.IsGuest {
		return "", custerr.NewForbiddenErr("guest users cannot import packs")
	}

	draft, err := s.parseSIQ(ctx, r, size)
	if err != nil {
		return "", err
	}
	draft.CreatedBy = user

	now := time.Now()
	draft.CreatedAt = now
	draft.UpdatedAt = now

	return s.packDraftRepo.Create(ctx, draft)
}
