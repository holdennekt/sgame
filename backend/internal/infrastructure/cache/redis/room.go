package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/pkg/custerr"
	"github.com/redis/go-redis/v9"
)

func getKey(id string) string {
	return domain.ROOM_PREFIX + id
}

func getLockKey(id string) string {
	return domain.LOCK_PREFIX + id
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

	pattern := fmt.Sprintf("%s*", domain.ROOM_PREFIX)
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

func (c *RoomCache) SafeSet(ctx context.Context, roomId string, updateFunc func(room *domain.Room) error) (*domain.Room, error) {
	lock, err := c.waitAndLock(ctx, roomId)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer lock.Release(ctx)

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
	if err := c.client.Del(ctx, getKey(roomId)).Err(); err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
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
		case <-time.After(100 * time.Millisecond):
		}
	}
}
