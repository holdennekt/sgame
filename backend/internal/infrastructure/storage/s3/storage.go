package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

// presigned URLs expire with the credentials that signed them, so cached role credentials keep at least this lifetime
const minCredentialsLifetime = time.Hour

type S3Storage struct {
	client     *awss3.Client
	presigner  *awss3.PresignClient
	uploader   *transfermanager.Client
	creds      aws.CredentialsProvider
	region     string
	bucketName string
	// off for the raw staging bucket, so nothing there is ever world-readable
	publicRead bool
	// path-style, so URLs contain "/<bucket>/" like GCS and Minio ones (cleanup-storage relies on it)
	publicBase string
	postURL    string
	userAgent  string
	httpClient *http.Client
}

// "public/" keys are readable through the bucket policy, not ACLs (disabled by default on S3).
func NewS3Storage(ctx context.Context, bucketName, userAgent string, publicRead bool) (*S3Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithCredentialsCacheOptions(func(o *aws.CredentialsCacheOptions) {
		o.ExpiryWindow = minCredentialsLifetime
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	if cfg.Region == "" {
		return nil, errors.New("AWS region is not set (AWS_REGION)")
	}

	client := awss3.NewFromConfig(cfg)
	if _, err := client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(bucketName)}); err != nil {
		return nil, fmt.Errorf("failed to access bucket %q: %w", bucketName, err)
	}

	return &S3Storage{
		client:     client,
		presigner:  awss3.NewPresignClient(client),
		uploader:   transfermanager.New(client),
		creds:      cfg.Credentials,
		region:     cfg.Region,
		bucketName: bucketName,
		publicRead: publicRead,
		publicBase: fmt.Sprintf("https://s3.%s.amazonaws.com/%s", cfg.Region, bucketName),
		postURL:    fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucketName, cfg.Region),
		userAgent:  userAgent,
		httpClient: &http.Client{
			Timeout: 60 * time.Minute,
		},
	}, nil
}

func (s *S3Storage) isPublicKey(key string) bool {
	return s.publicRead && strings.HasPrefix(key, "public/")
}

func isNotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return true
		}
	}
	return false
}

func escapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

func (s *S3Storage) Upload(ctx context.Context, ui storage.UploadInput) error {
	_, err := s.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(ui.Key),
		Body:        ui.Reader,
		ContentType: aws.String(ui.MimeType),
	})
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}
	return nil
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func (s *S3Storage) UploadFromURL(ctx context.Context, uui storage.URLUploadInput) error {
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

	body := &countingReader{r: io.LimitReader(resp.Body, uui.MaxBytes+1)}
	_, err = s.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(uui.Key),
		Body:        body,
		ContentType: aws.String(resp.Header.Get("Content-Type")),
	})
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to upload file: %w", err))
	}

	if body.n > uui.MaxBytes {
		_, _ = s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: aws.String(s.bucketName), Key: aws.String(uui.Key)})
		return custerr.NewBadRequestErr(fmt.Sprintf("file %q is too large. maximum size allowed: %d bytes", uui.URL, uui.MaxBytes))
	}

	return nil
}

func (s *S3Storage) Move(ctx context.Context, oldKey, newKey string) error {
	if err := s.Copy(ctx, oldKey, newKey); err != nil {
		return err
	}
	return s.Delete(ctx, oldKey)
}

func (s *S3Storage) Copy(ctx context.Context, oldKey, newKey string) error {
	_, err := s.client.CopyObject(ctx, &awss3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		Key:        aws.String(newKey),
		CopySource: aws.String(s.bucketName + "/" + escapeKey(oldKey)),
	})
	if err != nil {
		if isNotFound(err) {
			return custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", oldKey))
		}
		return custerr.NewInternalErr(fmt.Errorf("failed to copy file: %w", err))
	}
	return nil
}

// Delete checks existence first because S3 DeleteObject succeeds on missing keys.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if _, err := s.GetStats(ctx, key); err != nil {
		return err
	}
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return custerr.NewInternalErr(fmt.Errorf("failed to delete file: %w", err))
	}
	return nil
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to read file: %w", err))
	}
	return out.Body, nil
}

func (s *S3Storage) GetStats(ctx context.Context, key string) (*storage.Stats, error) {
	out, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return nil, custerr.NewInternalErr(fmt.Errorf("s3 stat error: %w", err))
	}
	return &storage.Stats{
		ContentType: aws.ToString(out.ContentType),
		Size:        aws.ToInt64(out.ContentLength),
		Checksum:    strings.Trim(aws.ToString(out.ETag), `"`),
	}, nil
}

func (s *S3Storage) SetContentType(ctx context.Context, key, contentType string) error {
	_, err := s.client.CopyObject(ctx, &awss3.CopyObjectInput{
		Bucket:            aws.String(s.bucketName),
		Key:               aws.String(key),
		CopySource:        aws.String(s.bucketName + "/" + escapeKey(key)),
		ContentType:       aws.String(contentType),
		MetadataDirective: types.MetadataDirectiveReplace,
	})
	if err != nil {
		if isNotFound(err) {
			return custerr.NewNotFoundErr(fmt.Sprintf("file with key %q not found", key))
		}
		return custerr.NewInternalErr(fmt.Errorf("failed to update content type: %w", err))
	}
	return nil
}

func (s *S3Storage) URL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if s.isPublicKey(key) {
		return s.publicBase + "/" + escapeKey(key), nil
	}

	req, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, awss3.WithPresignExpires(ttl))
	if err != nil {
		return "", custerr.NewInternalErr(fmt.Errorf("failed to generate presigned URL: %w", err))
	}
	return req.URL, nil
}

// SignUploadPolicy builds a SigV4 browser POST policy; the AWS SDK for Go v2 has no helper for it.
func (s *S3Storage) SignUploadPolicy(ctx context.Context, in storage.SignUploadPolicyInput) (*storage.SignUploadPolicyResult, error) {
	creds, err := s.creds.Retrieve(ctx)
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to retrieve AWS credentials: %w", err))
	}

	now := time.Now().UTC()
	date := now.Format("20060102")
	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", creds.AccessKeyID, date, s.region)

	fields := map[string]string{
		"key":              in.Key,
		"x-amz-algorithm":  "AWS4-HMAC-SHA256",
		"x-amz-credential": credential,
		"x-amz-date":       now.Format("20060102T150405Z"),
	}
	if creds.SessionToken != "" {
		fields["x-amz-security-token"] = creds.SessionToken
	}

	conditions := []any{
		map[string]string{"bucket": s.bucketName},
		[]any{"content-length-range", 0, in.MaxBytes},
	}
	for k, v := range fields {
		conditions = append(conditions, map[string]string{k: v})
	}

	policyJSON, err := json.Marshal(map[string]any{
		"expiration": now.Add(in.TTL).Format("2006-01-02T15:04:05.000Z"),
		"conditions": conditions,
	})
	if err != nil {
		return nil, custerr.NewInternalErr(fmt.Errorf("failed to encode upload policy: %w", err))
	}
	policy := base64.StdEncoding.EncodeToString(policyJSON)

	fields["policy"] = policy
	fields["x-amz-signature"] = hex.EncodeToString(hmacSHA256(signingKey(creds.SecretAccessKey, date, s.region, "s3"), policy))

	return &storage.SignUploadPolicyResult{
		URL:      s.postURL,
		FormData: fields,
	}, nil
}

func signingKey(secret, date, region, service string) []byte {
	key := hmacSHA256([]byte("AWS4"+secret), date)
	key = hmacSHA256(key, region)
	key = hmacSHA256(key, service)
	return hmacSHA256(key, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}
