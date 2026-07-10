package factory

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/queue/gcpubsub"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/queue/redisqueue"
	"github.com/holdennekt/sgame/backend/internal/interface/queue"
	"github.com/redis/go-redis/v9"
)

// New builds the notification pull-consumer matching cfg.StorageProvider:
// GCS uses a Pub/Sub pull subscription, everything else (Minio) uses a
// Redis-list-backed reliable queue fed by Minio's Redis notification target.
func New(ctx context.Context, cfg *config.Config, rds *redis.Client, maxConcurrent int) (queue.Consumer, error) {
	if cfg.StorageProvider == "gcs" {
		return gcpubsub.NewConsumer(ctx, cfg.GCPProjectID, cfg.PubSubSubscriptionID, maxConcurrent)
	}
	return redisqueue.NewConsumer(rds, maxConcurrent), nil
}
