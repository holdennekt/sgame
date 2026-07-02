package minio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client     *minio.Client
	bucketName string
	userAgent  string
	httpClient *http.Client
}

func NewMinioStorage(client *minio.Client, bucketName, userAgent string) storage.Storage {
	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			slog.Error("fatal error", "err", err)
			os.Exit(1)
		}
		slog.Info("bucket created", "bucket", bucketName)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "AllowPublicGet",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/public/*"]
			}
		]
	}`, bucketName)
	err = client.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}

	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
		userAgent:  userAgent,
		httpClient: &http.Client{
			Timeout: 60 * time.Minute,
		},
	}
}

func (s *MinioStorage) Upload(ctx context.Context, ui storage.UploadInput) error {
	_, err := s.client.PutObject(ctx, s.bucketName, ui.Key, ui.Reader, ui.Size, minio.PutObjectOptions{
		ContentType: ui.MimeType,
	})
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}

	return nil
}

func (s *MinioStorage) UploadFromURL(ctx context.Context, uui storage.URLUploadInput) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uui.URL, nil)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bts, _ := io.ReadAll(resp.Body)
		slog.Error("failed to get file from URL", "url", uui.URL, "body", string(bts))
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: unexpected status code %d for URL %s", resp.StatusCode, uui.URL))
	}

	if resp.ContentLength > uui.MaxBytes {
		return custerr.NewBadRequestErr(fmt.Sprintf("file %q is too large. maximum size allowed: %d bytes", uui.URL, uui.MaxBytes))
	}

	limitedReader := io.LimitReader(resp.Body, uui.MaxBytes+1)

	putRes, err := s.client.PutObject(
		ctx,
		s.bucketName,
		uui.Key,
		limitedReader,
		-1,
		minio.PutObjectOptions{
			ContentType: resp.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	if putRes.Size > uui.MaxBytes {
		_ = s.client.RemoveObject(ctx, s.bucketName, uui.Key, minio.RemoveObjectOptions{})
		return custerr.NewBadRequestErr(fmt.Sprintf("file \"%s\" is too large. maximum size allowed: %d bytes", uui.URL, uui.MaxBytes))
	}

	return nil
}

func (s *MinioStorage) Move(ctx context.Context, oldKey, newKey string) error {
	src := minio.CopySrcOptions{Bucket: s.bucketName, Object: oldKey}
	dst := minio.CopyDestOptions{Bucket: s.bucketName, Object: newKey}
	if _, err := s.client.CopyObject(ctx, dst, src); err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to copy file during move: %w", err))
	}
	return s.Delete(ctx, oldKey)
}

func (s *MinioStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return custerr.NewNotFoundErr(fmt.Sprintf("file with key \"%s\" not found", key))
		}
		return custerr.NewInternalErr(fmt.Errorf("failed to check file existence: %w", err))
	}

	err = s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to delete file from storage: %w", err))
	}

	return nil
}

func (s *MinioStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to read file: %w", err))
	}
	return obj, nil
}

func (s *MinioStorage) GetStats(ctx context.Context, key string) (*storage.Stats, error) {
	stats, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("minio stat error: %w", err))
	}
	return &storage.Stats{
		ContentType: stats.ContentType,
		Size:        stats.Size,
		Checksum:    stats.ChecksumSHA256,
	}, nil
}

func (s *MinioStorage) SetContentType(ctx context.Context, key, contentType string) error {
	src := minio.CopySrcOptions{Bucket: s.bucketName, Object: key}
	dst := minio.CopyDestOptions{
		Bucket:          s.bucketName,
		Object:          key,
		ContentType:     contentType,
		ReplaceMetadata: true,
	}
	if _, err := s.client.CopyObject(ctx, dst, src); err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return custerr.NewInternalErr(fmt.Errorf("failed to update content type: %w", err))
	}
	return nil
}

// URL returns a path relative to our own domain rather than an absolute URL —
// nginx proxies /storage/ on the same host:port the app is served on, so a
// relative URL resolves correctly regardless of which address (localhost,
// LAN IP, a tunnel, ...) the browser used to reach the app.
func (s *MinioStorage) URL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if strings.HasPrefix(key, "public/") {
		return fmt.Sprintf("/storage/%s/%s", s.bucketName, key), nil
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucketName, key, ttl, nil)
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to generate presigned URL: %w", err))
	}

	return "/storage" + u.Path + "?" + u.RawQuery, nil
}

func (s *MinioStorage) SignUploadPolicy(ctx context.Context, in storage.SignUploadPolicyInput) (*storage.SignUploadPolicyResult, error) {
	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(s.bucketName)
	_ = policy.SetKey(in.Key)
	_ = policy.SetExpires(time.Now().Add(in.TTL))
	_ = policy.SetContentLengthRange(0, in.MaxBytes)

	u, formData, err := s.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to generate presigned set URL and formData: %w", err))
	}

	return &storage.SignUploadPolicyResult{URL: "/storage" + u.Path, FormData: formData}, nil
}
