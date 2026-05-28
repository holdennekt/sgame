package storage

import (
	"context"
	"io"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
)

type Storage interface {
	Upload(ctx context.Context, ui UploadInput) error
	UploadFromURL(ctx context.Context, uui URLUploadInput) error
	Delete(ctx context.Context, key string) error
	Move(ctx context.Context, oldKey, newKey string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	GetStats(ctx context.Context, key string) (*Stats, error)
	URL(ctx context.Context, key string, ttl time.Duration) (string, error)
	SignUploadPolicy(ctx context.Context, in SignUploadPolicyInput) (*SignUploadPolicyResult, error)
}

type UploadInput struct {
	Key      string
	Size     int64
	MimeType string
	Type     domain.FileType
	Reader   io.Reader
}

type URLUploadInput struct {
	Key      string
	URL      string
	MaxBytes int64
}

type Stats struct {
	ContentType string
	Size        int64
	Checksum    string
}

type SignUploadPolicyInput struct {
	Key         string
	ContentType string
	TTL         time.Duration
	MaxBytes    int64
}

type SignUploadPolicyResult struct {
	URL      string
	FormData map[string]string
}
