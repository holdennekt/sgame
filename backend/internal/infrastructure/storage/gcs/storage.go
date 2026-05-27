package gcs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	gcsstorage "cloud.google.com/go/storage"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type GCSStorage struct {
	client     *gcsstorage.Client
	bucketName string
	publicBase string
	httpClient *http.Client
}

func NewGCSStorage(ctx context.Context, bucketName string) (storage.Storage, error) {
	client, err := gcsstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	_, err = client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to access bucket %q: %w", bucketName, err)
	}

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
		publicBase: "https://storage.googleapis.com",
		httpClient: &http.Client{
			Timeout: 60 * time.Minute,
		},
	}, nil
}

func (s *GCSStorage) Upload(ctx context.Context, ui storage.UploadInput) error {
	w := s.client.Bucket(s.bucketName).Object(ui.Key).NewWriter(ctx)
	w.ContentType = ui.MimeType

	if _, err := io.Copy(w, ui.Reader); err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	if err := w.Close(); err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to finalize upload: %w", err))
	}

	return nil
}

func (s *GCSStorage) UploadFromURL(ctx context.Context, uui storage.URLUploadInput) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uui.URL, nil)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bts, _ := io.ReadAll(resp.Body)
		log.Println("Failed to get file from URL", uui.URL, string(bts))
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: unexpected status code %d for URL %s", resp.StatusCode, uui.URL))
	}

	if resp.ContentLength > uui.MaxBytes {
		return custerr.NewBadRequestErr(fmt.Sprintf("file %q is too large. maximum size allowed: %d bytes", uui.URL, uui.MaxBytes))
	}

	w := s.client.Bucket(s.bucketName).Object(uui.Key).NewWriter(ctx)
	w.ContentType = resp.Header.Get("Content-Type")

	limitedReader := io.LimitReader(resp.Body, uui.MaxBytes+1)
	written, err := io.Copy(w, limitedReader)
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	if err := w.Close(); err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to finalize upload: %w", err))
	}

	if written > uui.MaxBytes {
		s.client.Bucket(s.bucketName).Object(uui.Key).Delete(ctx)
		return custerr.NewBadRequestErr(fmt.Sprintf("file %q is too large. maximum size allowed: %d bytes", uui.URL, uui.MaxBytes))
	}

	return nil
}

func (s *GCSStorage) Move(ctx context.Context, oldKey, newKey string) error {
	src := s.client.Bucket(s.bucketName).Object(oldKey)
	dst := s.client.Bucket(s.bucketName).Object(newKey)
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to copy file during move: %w", err))
	}
	return s.Delete(ctx, oldKey)
}

func (s *GCSStorage) Delete(ctx context.Context, key string) error {
	err := s.client.Bucket(s.bucketName).Object(key).Delete(ctx)
	if err != nil {
		if err == gcsstorage.ErrObjectNotExist {
			return custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return custerr.NewInternalErr(fmt.Errorf("failed to delete file: %w", err))
	}
	return nil
}

func (s *GCSStorage) GetStats(ctx context.Context, key string) (*storage.Stats, error) {
	attrs, err := s.client.Bucket(s.bucketName).Object(key).Attrs(ctx)
	if err != nil {
		if err == gcsstorage.ErrObjectNotExist {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return nil, custerr.NewInternalErr(fmt.Errorf("gcs stat error: %w", err))
	}
	return &storage.Stats{
		ContentType: attrs.ContentType,
		Size:        attrs.Size,
		Checksum:    fmt.Sprintf("%x", attrs.MD5),
	}, nil
}

func (s *GCSStorage) URL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if strings.HasPrefix(key, "public/") {
		return fmt.Sprintf("%s/%s/%s", s.publicBase, s.bucketName, key), nil
	}

	return s.signedURL(ctx, key, ttl)
}

func (s *GCSStorage) DirectURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return s.signedURL(ctx, key, ttl)
}

func (s *GCSStorage) signedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.Bucket(s.bucketName).SignedURL(key, &gcsstorage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(ttl),
		Scheme:  gcsstorage.SigningSchemeV4,
	})
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to generate signed URL: %w", err))
	}
	return u, nil
}

func (s *GCSStorage) SignUploadPolicy(ctx context.Context, in storage.SignUploadPolicyInput) (*storage.SignUploadPolicyResult, error) {
	conditions := []gcsstorage.PostPolicyV4Condition{
		gcsstorage.ConditionContentLengthRange(0, uint64(in.MaxBytes)),
	}

	opts := &gcsstorage.PostPolicyV4Options{
		Expires:    time.Now().Add(in.TTL),
		Conditions: conditions,
	}
	if strings.HasPrefix(in.Key, "public/") {
		opts.Fields = &gcsstorage.PolicyV4Fields{ACL: "public-read"}
	}

	policy, err := s.client.Bucket(s.bucketName).GenerateSignedPostPolicyV4(in.Key, opts)
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to generate upload policy: %w", err))
	}

	return &storage.SignUploadPolicyResult{
		URL:      policy.URL,
		FormData: policy.Fields,
	}, nil
}
