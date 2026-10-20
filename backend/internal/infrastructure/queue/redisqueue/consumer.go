package redisqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	queueKey      = "attachment:notifications:queue"
	processingKey = "attachment:notifications:processing"
	requeueDelay  = 5 * time.Second
)

type Consumer struct {
	rds     *redis.Client
	workers int
}

func NewConsumer(rds *redis.Client, workers int) *Consumer {
	return &Consumer{rds: rds, workers: workers}
}

func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, key string) error) error {
	if err := c.reclaim(ctx); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for range c.workers {
		wg.Go(func() { c.worker(ctx, handler) })
	}
	wg.Wait()
	return nil
}

// reclaim requeues what an ungraceful shutdown left in processingKey, ahead of newer messages.
func (c *Consumer) reclaim(ctx context.Context) error {
	for {
		err := c.rds.RPopLPush(ctx, processingKey, queueKey).Err()
		if errors.Is(err, redis.Nil) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (c *Consumer) worker(ctx context.Context, handler func(ctx context.Context, key string) error) {
	for {
		raw, err := c.rds.BLMove(ctx, queueKey, processingKey, "left", "right", 0).Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("redisqueue: failed to pop notification", "err", err)
			continue
		}

		keys, err := parseMinioEventKeys([]byte(raw))
		if err != nil {
			slog.Error("redisqueue: failed to parse notification, dropping", "err", err)
			c.rds.LRem(ctx, processingKey, 1, raw)
			continue
		}

		failed := false
		for _, key := range keys {
			if err := handler(ctx, key); err != nil {
				slog.Error("redisqueue: handler failed, will retry", "key", key, "err", err)
				failed = true
			}
		}

		if !failed {
			c.rds.LRem(ctx, processingKey, 1, raw)
			continue
		}

		select {
		case <-time.After(requeueDelay):
		case <-ctx.Done():
			return
		}
		c.rds.LRem(ctx, processingKey, 1, raw)
		c.rds.RPush(ctx, queueKey, raw)
	}
}

type minioRedisEvent struct {
	Event []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Event"`
}

func parseMinioEventKeys(raw []byte) ([]string, error) {
	var events []minioRedisEvent
	if err := json.Unmarshal(raw, &events); err != nil {
		return nil, fmt.Errorf("unrecognized notification payload: %w", err)
	}
	var keys []string
	for _, event := range events {
		for _, record := range event.Event {
			if !strings.Contains(record.EventName, "ObjectCreated") {
				continue
			}
			key, err := url.QueryUnescape(record.S3.Object.Key)
			if err != nil {
				return nil, fmt.Errorf("invalid object key %q: %w", record.S3.Object.Key, err)
			}
			keys = append(keys, key)
		}
	}
	return keys, nil
}
