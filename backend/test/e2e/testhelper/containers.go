package testhelper

import (
	"context"
	"fmt"
	"strings"

	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

type Containers struct {
	MongoURI  string
	RedisAddr string // "host:port"
	mongo     *tcmongo.MongoDBContainer
	redis     *tcredis.RedisContainer
}

func StartContainers(ctx context.Context) (*Containers, error) {
	mongoCtr, err := tcmongo.Run(ctx, "mongo:6.0")
	if err != nil {
		return nil, fmt.Errorf("start mongo: %w", err)
	}

	mongoURI, err := mongoCtr.ConnectionString(ctx)
	if err != nil {
		_ = mongoCtr.Terminate(ctx)
		return nil, fmt.Errorf("mongo connection string: %w", err)
	}

	// redis-stack-server bundles RedisJSON (JSONGet/JSONSet) and keyspace notifications
	redisCtr, err := tcredis.Run(ctx, "redis/redis-stack-server:7.4.0-v8")
	if err != nil {
		_ = mongoCtr.Terminate(ctx)
		return nil, fmt.Errorf("start redis: %w", err)
	}

	redisAddr, err := redisCtr.ConnectionString(ctx)
	if err != nil {
		_ = mongoCtr.Terminate(ctx)
		_ = redisCtr.Terminate(ctx)
		return nil, fmt.Errorf("redis connection string: %w", err)
	}
	// ConnectionString returns "redis://host:port"
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")

	return &Containers{
		MongoURI:  mongoURI,
		RedisAddr: redisAddr,
		mongo:     mongoCtr,
		redis:     redisCtr,
	}, nil
}

func (c *Containers) Terminate(ctx context.Context) {
	_ = c.mongo.Terminate(ctx)
	_ = c.redis.Terminate(ctx)
}
