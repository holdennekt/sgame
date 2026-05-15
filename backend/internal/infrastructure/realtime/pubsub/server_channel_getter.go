package pubsub

import (
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/redis/go-redis/v9"
)

type serverChannelGetter struct {
	client *redis.Client
}

func NewServerChannelGetter(client *redis.Client) realtime.ServerChannelGetter {
	return &serverChannelGetter{client}
}

func (g *serverChannelGetter) Get(name string) realtime.Channel {
	return NewChannel(g.client, name)
}
