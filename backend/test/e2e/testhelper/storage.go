package testhelper

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
)

type NoopStorage struct{}

func (s *NoopStorage) Upload(_ context.Context, _ storage.UploadInput) error { return nil }
func (s *NoopStorage) UploadFromURL(_ context.Context, _ storage.URLUploadInput) error { return nil }
func (s *NoopStorage) Delete(_ context.Context, _ string) error                        { return nil }
func (s *NoopStorage) Move(_ context.Context, _, _ string) error                       { return nil }
func (s *NoopStorage) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (s *NoopStorage) GetStats(_ context.Context, _ string) (*storage.Stats, error) {
	return nil, errors.New("not implemented")
}
func (s *NoopStorage) URL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "http://noop/" + key, nil
}
func (s *NoopStorage) SignUploadPolicy(_ context.Context, _ storage.SignUploadPolicyInput) (*storage.SignUploadPolicyResult, error) {
	return &storage.SignUploadPolicyResult{URL: "http://noop/upload", FormData: map[string]string{}}, nil
}
func (s *NoopStorage) SetContentType(_ context.Context, _, _ string) error { return nil }

var _ storage.Storage = (*NoopStorage)(nil)

// ensure domain import doesn't drift
var _ = domain.Image
