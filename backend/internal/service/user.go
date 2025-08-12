package service

import (
	"context"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/pkg/custerr"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository repository.User
}

func NewUserService(userRepository repository.User) *UserService {
	return &UserService{userRepository}
}

func (s *UserService) Create(ctx context.Context, user *domain.DbUser) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	user.Password = string(hashed)
	return s.userRepository.Create(ctx, user)
}

func (s *UserService) GetById(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user.User, nil
}

func (s *UserService) Update(ctx context.Context, user *domain.DbUser) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	user.Password = string(hashed)
	return s.userRepository.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.userRepository.Delete(ctx, id)
}
