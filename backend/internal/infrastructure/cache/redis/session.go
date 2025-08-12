package redis

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const SESSIONS_KEY = "sessions"

type SessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{client}
}

func (c *SessionCache) Get(ctx context.Context, sessionId string) (string, error) {
	userId, err := c.client.HGet(ctx, SESSIONS_KEY, sessionId).Result()
	if err != nil {
		if err == redis.Nil {
			return "", custerr.NewNotFoundErr("invalid session")
		}
		return "", custerr.NewInternalErr(err)
	}
	return userId, nil
}

func (c *SessionCache) GetKey(ctx context.Context, userId string) (string, error) {
	sessions, err := c.client.HGetAll(ctx, SESSIONS_KEY).Result()
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	for key, val := range sessions {
		if val == userId {
			return key, nil
		}
	}
	return "", custerr.NewNotFoundErr(fmt.Sprintf("no sessionId with value \"%s\"", userId))
}

func (c *SessionCache) Set(ctx context.Context, key string, val string) error {
	if err := c.client.HSet(ctx, SESSIONS_KEY, key, val).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
