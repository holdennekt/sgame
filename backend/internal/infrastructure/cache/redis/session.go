package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

const SESSION_KEY_PREFIX = "session:"
const SESSION_TTL = 7 * 24 * time.Hour

type SessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{client}
}

func (c *SessionCache) Get(ctx context.Context, sessionId string) (*domain.User, error) {
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

func (c *SessionCache) GetKey(ctx context.Context, userId string) (string, error) {
	keys, err := c.GetKeys(ctx, userId)
	if err != nil {
		return "", err
	}
	return keys[0], nil
}

func (c *SessionCache) GetKeys(ctx context.Context, userId string) ([]string, error) {
	var sessionIds []string
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, SESSION_KEY_PREFIX+"*", 100).Result()
		if err != nil {
			return nil, custerr.NewInternalErr(err)
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
				sessionIds = append(sessionIds, strings.TrimPrefix(key, SESSION_KEY_PREFIX))
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	if len(sessionIds) == 0 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no session for user \"%s\"", userId))
	}
	return sessionIds, nil
}

func (c *SessionCache) Set(ctx context.Context, sessionId string, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if err := c.client.Set(ctx, SESSION_KEY_PREFIX+sessionId, data, SESSION_TTL).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *SessionCache) Delete(ctx context.Context, sessionId string) error {
	if err := c.client.Del(ctx, SESSION_KEY_PREFIX+sessionId).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}
