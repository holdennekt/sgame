package pubsub

import "github.com/redis/go-redis/v9"

type ServerChannelGetter struct {
	client *redis.Client
}

func NewServerChannelGetter(client *redis.Client) *ServerChannelGetter {
	return &ServerChannelGetter{client}
}

func (g *ServerChannelGetter) Get(name string) any {
	return NewChannel(g.client, name)
}
