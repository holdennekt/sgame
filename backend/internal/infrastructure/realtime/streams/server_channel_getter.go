package streams

import (
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/redis/go-redis/v9"
)

type serverChannelGetter struct {
	client     *redis.Client
	persistent bool
}

func NewServerChannelGetter(client *redis.Client) realtime.ServerChannelGetter {
	return &serverChannelGetter{client: client, persistent: false}
}

func NewPersistentServerChannelGetter(client *redis.Client) realtime.ServerChannelGetter {
	return &serverChannelGetter{client: client, persistent: true}
}

func (g *serverChannelGetter) Get(name string) realtime.Channel {
	return NewChannel(g.client, name, g.persistent)
}
