package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const SESSION_KEY_PREFIX = "session:"
const SESSION_TTL = 7 * 24 * time.Hour

type sessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) cache.Session {
	return &sessionCache{client}
}

func (c *sessionCache) Get(ctx context.Context, sessionId string) (*domain.User, error) {
	raw, err := c.client.Get(ctx, SESSION_KEY_PREFIX+sessionId).Result()
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
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, SESSION_KEY_PREFIX+"*", 100).Result()
		if err != nil {
			return "", custerr.NewInternalErr(err)
		}
		for _, key := range keys {
			raw, err := c.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			var user domain.User
			if err := json.Unmarshal([]byte(raw), &user); err != nil {
				continue
			}
			if user.Id == userId {
				return strings.TrimPrefix(key, SESSION_KEY_PREFIX), nil
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return "", custerr.NewNotFoundErr(fmt.Sprintf("no session for user \"%s\"", userId))
}

func (c *sessionCache) Set(ctx context.Context, sessionId string, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Set(ctx, SESSION_KEY_PREFIX+sessionId, data, SESSION_TTL).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *sessionCache) Delete(ctx context.Context, sessionId string) error {
	if err := c.client.Del(ctx, SESSION_KEY_PREFIX+sessionId).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
