package factory

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/config"
	gcsStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/gcs"
	minioStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/minio"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func New(ctx context.Context, cfg *config.Config, bucketName string, publicRead bool) (storage.Storage, error) {
	if cfg.StorageProvider == "gcs" {
		return gcsStorage.NewGCSStorage(ctx, bucketName, publicRead)
	}

	mio, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.MinioUser,
			cfg.MinioPassword,
			"",
		),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}
	return minioStorage.NewMinioStorage(mio, bucketName, cfg.UserAgent, publicRead), nil
}
