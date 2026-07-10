package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/vansante/go-ffprobe.v2"
)

const AVATAR_DIMENSION = 256

type UserService struct {
	userRepository    repository.User
	sessionCache      cache.Session
	compressedStorage storage.Compressed
}

func NewUserService(userRepository repository.User, sessionCache cache.Session, compressedStorage storage.Compressed) *UserService {
	return &UserService{userRepository, sessionCache, compressedStorage}
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
	if sessionIds, err := s.sessionCache.GetKeys(ctx, user.Id); err == nil {
		for _, sessionId := range sessionIds {
			if err := s.sessionCache.Set(ctx, sessionId, &user.User); err != nil {
				slog.Warn("failed to update session cache", "err", err)
			}
		}
	}
	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userId string, r io.Reader) (string, error) {
	existing, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return "", err
	}

	inFile, err := os.CreateTemp("", "sgame-avatar-in-*")
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer func() { _ = os.Remove(inFile.Name()) }()
	defer func() { _ = inFile.Close() }()
	if _, err := io.Copy(inFile, r); err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to buffer avatar: %w", err))
	}

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	probeData, probeErr := ffprobe.ProbeURL(probeCtx, inFile.Name())
	cancel()
	if probeErr != nil {
		return "", custerr.NewBadRequestErr("avatar must be a valid image")
	}
	if fileType, _ := classifyProbe(probeData); fileType != domain.Image {
		return "", custerr.NewBadRequestErr("avatar must be an image")
	}

	outFile, err := os.CreateTemp("", "sgame-avatar-out-*")
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer func() { _ = os.Remove(outFile.Name()) }()
	defer func() { _ = outFile.Close() }()

	if err := transcodeAvatar(ctx, inFile.Name(), outFile.Name(), AVATAR_DIMENSION); err != nil {
		var ce ContentError
		if errors.As(err, &ce) {
			return "", custerr.NewBadRequestErr("unsupported or corrupt image, try a different file")
		}
		return "", custerr.NewInternalErr(err)
	}

	stat, err := outFile.Stat()
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to stat avatar: %w", err))
	}
	out, err := os.Open(outFile.Name())
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to open avatar: %w", err))
	}
	defer func() { _ = out.Close() }()

	key := "public/" + uuid.NewString()
	if err := s.compressedStorage.Upload(ctx, storage.UploadInput{
		Key:      key,
		Size:     stat.Size(),
		MimeType: IMAGE_MIME_TYPE,
		Type:     domain.Image,
		Reader:   out,
	}); err != nil {
		return "", err
	}

	url, err := s.compressedStorage.URL(ctx, key, GET_URL_TTL)
	if err != nil {
		return "", err
	}

	if err := s.Update(ctx, &domain.DbUser{
		User: domain.User{Id: userId, Name: existing.Name, Avatar: &url},
	}); err != nil {
		return "", err
	}
	return url, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	if err := s.userRepository.Delete(ctx, id); err != nil {
		return err
	}
	if sessionIds, err := s.sessionCache.GetKeys(ctx, id); err == nil {
		for _, sessionId := range sessionIds {
			if err := s.sessionCache.Delete(ctx, sessionId); err != nil {
				slog.Warn("failed to delete session from cache", "err", err)
			}
		}
	}
	return nil
}
