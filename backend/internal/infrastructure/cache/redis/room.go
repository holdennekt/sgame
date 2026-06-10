package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/bsm/redislock"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/transport/http"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

func getKey(id string) string {
	return domain.ROOM_PREFIX + id + domain.STATE_POSTFIX
}

func getLockKey(id string) string {
	return domain.ROOM_PREFIX + id + domain.STATE_POSTFIX + domain.LOCK_POSTFIX
}

type roomCache struct {
	client *redis.Client
	locker *redislock.Client
}

func NewRoomCache(client *redis.Client) cache.Room {
	return &roomCache{client: client, locker: redislock.New(client)}
}

func (c *roomCache) getByKey(ctx context.Context, key string) (*domain.Room, error) {
	var room domain.Room

	res, err := c.client.JSONGet(ctx, key).Result()
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	if res == "" {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no key \"%s\"", key))
	}

	if err := json.Unmarshal([]byte(res), &room); err != nil {
		return nil, custerr.NewInternalErr(err)
	}

	return &room, nil
}

func (c *roomCache) GetById(ctx context.Context, id string) (*domain.Room, error) {
	return c.getByKey(ctx, getKey(id))
}

func (c *roomCache) Get(ctx context.Context) ([]domain.RoomLobby, error) {
	roomLobbys := make([]domain.RoomLobby, 0)

	pattern := fmt.Sprintf("%s*%s", domain.ROOM_PREFIX, domain.STATE_POSTFIX)
	iter := c.client.Scan(ctx, uint64(c.client.Options().DB), pattern, 10).Iterator()
	for iter.Next(ctx) {
		room, err := c.getByKey(ctx, iter.Val())
		if err != nil {
			return nil, err
		}

		roomLobbys = append(roomLobbys, domain.NewRoomLobby(room))
	}
	return roomLobbys, nil
}

func (c *roomCache) Set(ctx context.Context, room *domain.Room) error {
	if err := c.client.JSONSet(ctx, getKey(room.Id), "$", room).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *roomCache) SafeUpdate(ctx context.Context, roomId string, updateFunc func(room *domain.Room) error) (*domain.Room, error) {
	lock, err := c.waitAndLock(ctx, roomId)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer func() { _ = lock.Release(ctx) }()

	room, err := c.GetById(ctx, roomId)
	if err != nil {
		return nil, err
	}

	if err := updateFunc(room); err != nil {
		return nil, err
	}

	if err := c.client.JSONSet(ctx, getKey(room.Id), "$", room).Err(); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return room, nil
}

func (c *roomCache) Delete(ctx context.Context, roomId string) error {
	if err := c.client.Del(ctx, getKey(roomId), getLockKey(roomId)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *roomCache) Expire(ctx context.Context, roomId string, duration time.Duration) error {
	if err := c.client.Expire(ctx, getKey(roomId), duration).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *roomCache) Persist(ctx context.Context, roomId string) error {
	if err := c.client.Persist(ctx, getKey(roomId)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *roomCache) TrySetOwner(ctx context.Context, roomId string, ttl time.Duration) (bool, error) {
	ip := http.GetServerIP()
	key := domain.ROOM_PREFIX + roomId + domain.INTERNAL_POSTFIX + domain.OWNER_POSTFIX
	return c.client.SetNX(ctx, key, ip, ttl).Result()
}

func (c *roomCache) UpdateOwner(ctx context.Context, roomId string, ttl time.Duration) error {
	key := domain.ROOM_PREFIX + roomId + domain.INTERNAL_POSTFIX + domain.OWNER_POSTFIX
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *roomCache) ListenForExpiredOwners(ctx context.Context, handleExpiredOwner func(roomId string)) {
	pubsubExp := c.client.PSubscribe(ctx, "__keyevent@0__:expired")
	defer func() { _ = pubsubExp.Close() }()

	re := regexp.MustCompile(fmt.Sprintf("^%s(.+)%s%s$", domain.ROOM_PREFIX, domain.INTERNAL_POSTFIX, domain.OWNER_POSTFIX))
	ch := pubsubExp.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			matches := re.FindStringSubmatch(msg.Payload)
			if len(matches) != 2 {
				continue
			}
			id := matches[1]
			slog.Debug("owner key expired", "room_id", id)
			handleExpiredOwner(id)
		}
	}
}

func (c *roomCache) waitAndLock(ctx context.Context, roomId string) (*redislock.Lock, error) {
	for {
		lock, err := c.locker.Obtain(ctx, getLockKey(roomId), time.Second, nil)
		if err == nil {
			return lock, nil
		}

		if err != redislock.ErrNotObtained {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
