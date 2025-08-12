package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/pkg/custerr"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	sessionCache   cache.Session
	userRepository repository.User
}

func NewAuthService(sessionCache cache.Session, userRepository repository.User) *AuthService {
	return &AuthService{sessionCache, userRepository}
}

func (s *AuthService) Login(ctx context.Context, dto domain.DbUserDTO) (sessionId string, userId string, err error) {
	dbUser, err := s.userRepository.GetByLogin(ctx, dto.Login)
	if err != nil {
		return
	}
	userId = dbUser.Id

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(dto.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			err = custerr.NewUnauthorizedErr("wrong password")
			return
		}
		err = custerr.NewInternalErr(err)
		return
	}

	sessionId, err = s.sessionCache.GetKey(ctx, userId)
	if err == nil {
		return
	}
	if _, ok := err.(custerr.InternalErr); ok {
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	sessionId = id.String()
	err = s.sessionCache.Set(ctx, sessionId, dbUser.Id)
	return
}

func (s *AuthService) Register(ctx context.Context, dto domain.DbUserDTO) (sessionId string, userId string, err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	dto.Password = string(hashed)

	dbUser := &domain.DbUser{
		User: domain.User{
			Name: dto.Login,
		},
		DbUserDTO: dto,
	}

	userId, err = s.userRepository.Create(ctx, dbUser)
	if err != nil {
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	sessionId = id.String()
	err = s.sessionCache.Set(ctx, sessionId, userId)
	return
}

func (s *AuthService) GetUserID(ctx context.Context, sessionId string) (string, error) {
	return s.sessionCache.Get(ctx, sessionId)
}
