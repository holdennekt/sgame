package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	sessionCache   cache.Session
	userRepository repository.User
}

func NewAuthService(sessionCache cache.Session, userRepository repository.User) *AuthService {
	return &AuthService{sessionCache, userRepository}
}

func (s *AuthService) Login(ctx context.Context, cur dto.CreateUserRequest) (sessionId string, userId string, err error) {
	dbUser, err := s.userRepository.GetByLogin(ctx, cur.Login)
	if err != nil {
		return
	}
	userId = dbUser.Id

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(cur.Password))
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
	err = s.sessionCache.Set(ctx, sessionId, &dbUser.User)
	return
}

func (s *AuthService) Register(ctx context.Context, cur dto.CreateUserRequest) (sessionId string, userId string, err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(cur.Password), bcrypt.DefaultCost)
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}

	dbUser := &domain.DbUser{
		User: domain.User{
			Name: cur.Login,
		},
		Login:    cur.Login,
		Password: string(hashed),
	}

	userId, err = s.userRepository.Create(ctx, dbUser)
	if err != nil {
		return
	}
	dbUser.User.Id = userId

	id, err := uuid.NewRandom()
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	sessionId = id.String()
	err = s.sessionCache.Set(ctx, sessionId, &dbUser.User)
	return
}

func (s *AuthService) GuestLogin(ctx context.Context, name string) (sessionId string, userId string, err error) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	userId = userUUID.String()

	guest := &domain.User{
		Id:      userId,
		Name:    name,
		IsGuest: true,
	}

	id, err := uuid.NewRandom()
	if err != nil {
		err = custerr.NewInternalErr(err)
		return
	}
	sessionId = id.String()
	err = s.sessionCache.Set(ctx, sessionId, guest)
	return
}

func (s *AuthService) GetUser(ctx context.Context, sessionId string) (*domain.User, error) {
	user, err := s.sessionCache.Get(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionId string) error {
	return s.sessionCache.Delete(ctx, sessionId)
}
