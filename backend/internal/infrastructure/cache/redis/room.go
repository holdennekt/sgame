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

func getSpectatorsKey(id string) string {
	return domain.ROOM_PREFIX + id + domain.SPECTATORS_POSTFIX
}

func getOwnerKey(id string) string {
	return domain.ROOM_PREFIX + id + domain.INTERNAL_POSTFIX + domain.OWNER_POSTFIX
}

type RoomCache struct {
	client *redis.Client
	locker *redislock.Client
}

func NewRoomCache(client *redis.Client) *RoomCache {
	return &RoomCache{client: client, locker: redislock.New(client)}
}

func (c *RoomCache) getByKey(ctx context.Context, key string) (*domain.Room, error) {
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

func (c *RoomCache) GetById(ctx context.Context, id string) (*domain.Room, error) {
	return c.getByKey(ctx, getKey(id))
}

func (c *RoomCache) Get(ctx context.Context) ([]domain.RoomLobby, error) {
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

func (c *RoomCache) Set(ctx context.Context, room *domain.Room) error {
	if err := c.client.JSONSet(ctx, getKey(room.Id), "$", room).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *RoomCache) SafeUpdate(ctx context.Context, roomId string, updateFunc func(room *domain.Room) error) (*domain.Room, error) {
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

func (c *RoomCache) Delete(ctx context.Context, roomId string) error {
	if err := c.client.Del(ctx, getKey(roomId), getLockKey(roomId), getSpectatorsKey(roomId), getOwnerKey(roomId)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *RoomCache) IncrSpectators(ctx context.Context, roomId string) (int, error) {
	n, err := c.client.Incr(ctx, getSpectatorsKey(roomId)).Result()
	return int(n), err
}

func (c *RoomCache) DecrSpectators(ctx context.Context, roomId string) (int, error) {
	n, err := c.client.Decr(ctx, getSpectatorsKey(roomId)).Result()
	return int(n), err
}

func (c *RoomCache) GetSpectatorCount(ctx context.Context, roomId string) (int, error) {
	n, err := c.client.Get(ctx, getSpectatorsKey(roomId)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

func (c *RoomCache) Expire(ctx context.Context, roomId string, duration time.Duration) error {
	if err := c.client.Expire(ctx, getKey(roomId), duration).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *RoomCache) Persist(ctx context.Context, roomId string) error {
	if err := c.client.Persist(ctx, getKey(roomId)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (c *RoomCache) TrySetOwner(ctx context.Context, roomId string, ttl time.Duration) (bool, error) {
	ip := http.GetServerIP()
	return c.client.SetNX(ctx, getOwnerKey(roomId), ip, ttl).Result()
}

func (c *RoomCache) UpdateOwner(ctx context.Context, roomId string, ttl time.Duration) error {
	key := getOwnerKey(roomId)
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *RoomCache) ListenForExpiredOwners(ctx context.Context, handleExpiredOwner func(roomId string)) {
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

func (c *RoomCache) waitAndLock(ctx context.Context, roomId string) (*redislock.Lock, error) {
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
