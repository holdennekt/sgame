package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	waitTimeSeconds = 20
	// failed messages reappear after this delay; the redrive policy caps attempts
	retryVisibilitySeconds = 10
)

type Consumer struct {
	client   *awssqs.Client
	queueURL string
	workers  int
}

func NewConsumer(ctx context.Context, queueURL string, workers int) (*Consumer, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return &Consumer{client: awssqs.NewFromConfig(cfg), queueURL: queueURL, workers: workers}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, key string) error) error {
	var wg sync.WaitGroup
	for range c.workers {
		wg.Go(func() { c.worker(ctx, handler) })
	}
	wg.Wait()
	return nil
}

func (c *Consumer) worker(ctx context.Context, handler func(ctx context.Context, key string) error) {
	for {
		out, err := c.client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     waitTimeSeconds,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("sqs: failed to receive notification", "err", err)
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, msg := range out.Messages {
			c.process(ctx, msg, handler)
		}
	}
}

func (c *Consumer) process(ctx context.Context, msg types.Message, handler func(ctx context.Context, key string) error) {
	keys, err := parseS3EventKeys([]byte(aws.ToString(msg.Body)))
	if err != nil {
		slog.Error("sqs: failed to parse notification, dropping", "err", err)
		c.delete(ctx, msg)
		return
	}

	for _, key := range keys {
		if err := handler(ctx, key); err != nil {
			slog.Error("sqs: handler failed, will retry", "key", key, "err", err)
			_, _ = c.client.ChangeMessageVisibility(ctx, &awssqs.ChangeMessageVisibilityInput{
				QueueUrl:          aws.String(c.queueURL),
				ReceiptHandle:     msg.ReceiptHandle,
				VisibilityTimeout: retryVisibilitySeconds,
			})
			return
		}
	}
	c.delete(ctx, msg)
}

func (c *Consumer) delete(ctx context.Context, msg types.Message) {
	_, err := c.client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil && ctx.Err() == nil {
		slog.Error("sqs: failed to delete notification", "err", err)
	}
}

type s3Event struct {
	Records []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// parseS3EventKeys returns no keys for the s3:TestEvent S3 sends when the notification is set up.
func parseS3EventKeys(raw []byte) ([]string, error) {
	var event s3Event
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, fmt.Errorf("unrecognized notification payload: %w", err)
	}
	var keys []string
	for _, record := range event.Records {
		if !strings.HasPrefix(record.EventName, "ObjectCreated") {
			continue
		}
		key, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			return nil, fmt.Errorf("invalid object key %q: %w", record.S3.Object.Key, err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}
