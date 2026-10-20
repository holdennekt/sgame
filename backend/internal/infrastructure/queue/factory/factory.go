package factory

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/queue/gcpubsub"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/queue/redisqueue"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/queue/sqs"
	"github.com/holdennekt/sgame/backend/internal/interface/queue"
	"github.com/redis/go-redis/v9"
)

// New picks the upload-notification consumer: Pub/Sub (gcs), SQS (s3), or a Redis list fed by Minio (anything else).
func New(ctx context.Context, cfg *config.Config, rds *redis.Client, maxConcurrent int) (queue.Consumer, error) {
	switch cfg.StorageProvider {
	case "gcs":
		return gcpubsub.NewConsumer(ctx, cfg.GCPProjectID, cfg.PubSubSubscriptionID, maxConcurrent)
	case "s3":
		return sqs.NewConsumer(ctx, cfg.SQSQueueURL, maxConcurrent)
	}
	return redisqueue.NewConsumer(rds, maxConcurrent), nil
}
