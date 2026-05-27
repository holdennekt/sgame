package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const SESSIONS_KEY = "sessions"

type sessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) cache.Session {
	return &sessionCache{client}
}

func (c *sessionCache) Get(ctx context.Context, sessionId string) (*domain.User, error) {
	raw, err := c.client.HGet(ctx, SESSIONS_KEY, sessionId).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, custerr.NewNotFoundErr("invalid session")
		}
		return nil, custerr.NewInternalErr(err)
	}
	var user domain.User
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return &user, nil
}

func (c *sessionCache) GetKey(ctx context.Context, userId string) (string, error) {
	sessions, err := c.client.HGetAll(ctx, SESSIONS_KEY).Result()
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	for sessionId, raw := range sessions {
		var user domain.User
		if err := json.Unmarshal([]byte(raw), &user); err != nil {
			continue
		}
		if user.Id == userId {
			return sessionId, nil
		}
	}
	return "", custerr.NewNotFoundErr(fmt.Sprintf("no session for user \"%s\"", userId))
}

func (c *sessionCache) Set(ctx context.Context, sessionId string, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.HSet(ctx, SESSIONS_KEY, sessionId, data).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *sessionCache) Delete(ctx context.Context, sessionId string) error {
	if err := c.client.HDel(ctx, SESSIONS_KEY, sessionId).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
