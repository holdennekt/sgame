package service

import (
	"context"
	"crypto/sha1"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/pkg/custerr"
)

type PackService struct {
	userRepository repository.User
	packRepository repository.Pack
}

func NewPackService(userRepository repository.User, packRepository repository.Pack) *PackService {
	return &PackService{userRepository, packRepository}
}

func (s *PackService) validateRoundsCheckSum(ctx context.Context, packId string, packDTO *domain.PackDTO) ([]byte, error) {
	marshaledRounds, _ := json.Marshal(struct {
		Rounds     []domain.Round    `json:"rounds"`
		FinalRound domain.FinalRound `json:"finalRound"`
	}{Rounds: packDTO.Rounds, FinalRound: packDTO.FinalRound})
	hasher := sha1.New()
	_, err := hasher.Write(marshaledRounds)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	roundsChecksum, err := hasher.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}

	pack, err := s.packRepository.GetByRoundsChecksum(
		ctx,
		dto.GetPackByRoundsChecksumDTO{RoundsChecksum: roundsChecksum, IgnoreId: packId},
	)
	if err != nil {
		if _, ok := err.(custerr.NotFoundErr); ok {
			return roundsChecksum, nil
		}
		return nil, err
	}
	return nil, custerr.NewConflictErr(
		fmt.Sprintf("the pack with such rounds already exists and has id \"%s\"", pack.Id),
	)
}

func (s *PackService) Create(ctx context.Context, dto dto.CreatePackDTO) (string, error) {
	user, err := s.userRepository.GetById(ctx, dto.UserId)
	if err != nil {
		return "", err
	}

	content := []string{dto.PackDTO.Name}
	for _, round := range dto.PackDTO.Rounds {
		for _, category := range round.Categories {
			content = append(content, category.Name)
		}
		content = append(content, round.Name)
	}

	roundsCheckSum, err := s.validateRoundsCheckSum(ctx, "", dto.PackDTO)
	if err != nil {
		return "", err
	}

	pack := &domain.Pack{
		CreatedBy:      user.User,
		RoundsChecksum: roundsCheckSum,
		Content:        strings.Join(content, ", "),
		PackDTO:        *dto.PackDTO,
	}

	return s.packRepository.Create(ctx, pack)
}

func (s *PackService) GetById(ctx context.Context, dto dto.GetPackByIdDTO) (*domain.Pack, error) {
	pack, err := s.packRepository.GetById(ctx, dto.Id)
	if err != nil {
		return nil, err
	}
	if pack.Type != "public" && dto.UserId != pack.CreatedBy.Id && dto.UserId != "" {
		return nil, custerr.NewForbiddenErr("can not get private pack")
	}
	return pack, nil
}

func (s *PackService) GetPreviews(ctx context.Context, dto dto.GetPacksDTO) ([]domain.PackPreview, error) {
	return s.packRepository.GetPreviews(ctx, dto)
}

func (s *PackService) GetHiddens(ctx context.Context, dto dto.GetPacksDTO) ([]domain.HiddenPack, int, error) {
	packs, err := s.packRepository.GetHiddens(ctx, dto)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.packRepository.GetCount(ctx, dto)
	if err != nil {
		return nil, 0, err
	}
	return packs, count, nil
}

func (s *PackService) Update(ctx context.Context, dto dto.UpdatePackDTO) error {
	pack, err := s.packRepository.GetById(ctx, dto.Id)
	if err != nil {
		return err
	}

	if pack.CreatedBy.Id != dto.UserId {
		return custerr.NewForbiddenErr("can only edit your own packs")
	}

	content := []string{dto.PackDTO.Name}
	for _, round := range dto.PackDTO.Rounds {
		for _, category := range round.Categories {
			content = append(content, category.Name)
		}
		content = append(content, round.Name)
	}

	newRoundsCheckSum, err := s.validateRoundsCheckSum(ctx, dto.Id, dto.PackDTO)
	if err != nil {
		return err
	}

	pack.RoundsChecksum = newRoundsCheckSum
	pack.Content = strings.Join(content, ", ")
	pack.PackDTO = *dto.PackDTO

	return s.packRepository.Update(ctx, pack)
}

func (s *PackService) Delete(ctx context.Context, dto dto.DeletePackDTO) error {
	pack, err := s.packRepository.GetById(ctx, dto.Id)
	if err != nil {
		return err
	}

	if pack.CreatedBy.Id != dto.UserId {
		return custerr.NewForbiddenErr("can only delete your own packs")
	}

	return s.packRepository.Delete(ctx, dto.Id)
}
