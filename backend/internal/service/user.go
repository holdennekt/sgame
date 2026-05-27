package service

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository repository.User
	sessionCache   cache.Session
}

func NewUserService(userRepository repository.User, sessionCache cache.Session) *UserService {
	return &UserService{userRepository, sessionCache}
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
		if _, ok := err.(custerr.NotFoundErr); ok {
			sessionId, err := s.sessionCache.GetKey(ctx, id)
			if err != nil {
				return nil, custerr.NewNotFoundErr("user not found")
			}
			return s.sessionCache.Get(ctx, sessionId)
		}
		return nil, err
	}
	return &user.User, nil
}

func (s *UserService) Update(ctx context.Context, user *domain.DbUser) error {
	existing, err := s.userRepository.GetById(ctx, user.Id)
	if err != nil {
		return err
	}
	user.Login = existing.Login
	if user.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return custerr.NewInternalErr(err)
		}
		user.Password = string(hashed)
	} else {
		user.Password = existing.Password
	}
	if user.Avatar != nil && *user.Avatar == "" {
		user.Avatar = nil
	}
	if err := s.userRepository.Update(ctx, user); err != nil {
		return err
	}
	if sessionId, err := s.sessionCache.GetKey(ctx, user.Id); err == nil {
		s.sessionCache.Set(ctx, sessionId, &user.User)
	}
	return nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	if err := s.userRepository.Delete(ctx, id); err != nil {
		return err
	}
	if sessionId, err := s.sessionCache.GetKey(ctx, id); err == nil {
		s.sessionCache.Delete(ctx, sessionId)
	}
	return nil
}
