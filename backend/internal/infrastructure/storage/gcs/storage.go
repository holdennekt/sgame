package gcs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	iampb "cloud.google.com/go/iam/apiv1/iampb"
	gcsstorage "cloud.google.com/go/storage"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"google.golang.org/genproto/googleapis/type/expr"
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

	if err := ensurePublicPrefixPolicy(ctx, client.Bucket(bucketName), bucketName); err != nil {
		return nil, err
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

func ensureUniformAccess(ctx context.Context, bucket *gcsstorage.BucketHandle) error {
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bucket attrs: %w", err)
	}
	if attrs.UniformBucketLevelAccess.Enabled {
		return nil
	}
	_, err = bucket.Update(ctx, gcsstorage.BucketAttrsToUpdate{
		UniformBucketLevelAccess: &gcsstorage.UniformBucketLevelAccess{Enabled: true},
	})
	if err != nil {
		return fmt.Errorf("failed to enable uniform bucket-level access: %w", err)
	}
	log.Printf("Enabled uniform bucket-level access on bucket")
	return nil
}

func ensurePublicPrefixPolicy(ctx context.Context, bucket *gcsstorage.BucketHandle, bucketName string) error {
	if err := ensureUniformAccess(ctx, bucket); err != nil {
		return err
	}

	policy, err := bucket.IAM().V3().Policy(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bucket IAM policy: %w", err)
	}

	conditionExpr := fmt.Sprintf(`resource.name.startsWith("projects/_/buckets/%s/objects/public/")`, bucketName)

	for _, b := range policy.Bindings {
		if b.Role == "roles/storage.objectViewer" && b.Condition != nil && b.Condition.Expression == conditionExpr && slices.Contains(b.Members, "allUsers") {
			return nil
		}
	}

	policy.Bindings = append(policy.Bindings, &iampb.Binding{
		Role:    "roles/storage.objectViewer",
		Members: []string{"allUsers"},
		Condition: &expr.Expr{
			Expression: conditionExpr,
			Title:      "public-prefix-read",
		},
	})

	if err := bucket.IAM().V3().SetPolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to set bucket IAM policy: %w", err)
	}

	log.Printf("Granted public read on gs://%s/public/*", bucketName)
	return nil
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

	policy, err := s.client.Bucket(s.bucketName).GenerateSignedPostPolicyV4(in.Key, &gcsstorage.PostPolicyV4Options{
		Expires:    time.Now().Add(in.TTL),
		Conditions: conditions,
	})
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to generate upload policy: %w", err))
	}

	return &storage.SignUploadPolicyResult{
		URL:      policy.URL,
		FormData: policy.Fields,
	}, nil
}
