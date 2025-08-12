//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	redisCache "github.com/holdennekt/sgame/internal/infrastructure/cache/redis"
	mongoDatabase "github.com/holdennekt/sgame/internal/infrastructure/database/mongo"
	"github.com/holdennekt/sgame/internal/infrastructure/realtime/pubsub"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/service"
	"github.com/holdennekt/sgame/internal/transport/http"
	"github.com/holdennekt/sgame/internal/transport/ws"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserRepoSet = wire.NewSet(
	mongoDatabase.NewUserRepository,
	wire.Bind(new(repository.User), new(*mongoDatabase.UserRepository)),
)

var PackRepoSet = wire.NewSet(
	mongoDatabase.NewPackRepository,
	wire.Bind(new(repository.Pack), new(*mongoDatabase.PackRepository)),
)

var RoomRepoSet = wire.NewSet(
	mongoDatabase.NewRoomRepository,
	wire.Bind(new(repository.Room), new(*mongoDatabase.RoomRepository)),
)

var SessionCacheSet = wire.NewSet(
	redisCache.NewSessionCache,
	wire.Bind(new(cache.Session), new(*redisCache.SessionCache)),
)

var RoomCacheSet = wire.NewSet(
	redisCache.NewRoomCache,
	wire.Bind(new(cache.Room), new(*redisCache.RoomCache)),
)

var ServerChannelGetterSet = wire.NewSet(
	pubsub.NewServerChannelGetter,
	wire.Bind(new(realtime.ServerChannelGetter), new(*pubsub.ServerChannelGetter)),
)

func InitializeAuthController(mdb *mongo.Database, rds *redis.Client) *http.AuthController {
	wire.Build(
		SessionCacheSet,
		UserRepoSet,
		service.NewAuthService,
		http.NewAuthController,
	)
	return nil
}

func InitializeUserController(mdb *mongo.Database) *http.UserController {
	wire.Build(UserRepoSet, service.NewUserService, http.NewUserController)
	return nil
}

func InitializePackController(mdb *mongo.Database) *http.PackController {
	wire.Build(UserRepoSet, PackRepoSet, service.NewPackService, http.NewPackController)
	return nil
}

func InitializeRoomController(mdb *mongo.Database, rds *redis.Client) *http.RoomController {
	wire.Build(
		UserRepoSet,
		PackRepoSet,
		RoomRepoSet,
		RoomCacheSet,
		ServerChannelGetterSet,
		service.NewPackService,
		service.NewRoomService,
		http.NewRoomController,
	)
	return nil
}

func InitializeLobbyHandler(mdb *mongo.Database, rds *redis.Client) *ws.LobbyHandler {
	wire.Build(
		ServerChannelGetterSet,
		RoomCacheSet,
		UserRepoSet,
		service.NewUserService,
		ws.NewLobbyHandler,
	)
	return nil
}

func InitializeRoomHandler(mdb *mongo.Database, rds *redis.Client) *ws.RoomHandler {
	wire.Build(
		ServerChannelGetterSet,
		RoomCacheSet,
		UserRepoSet,
		PackRepoSet,
		RoomRepoSet,
		service.NewUserService,
		service.NewPackService,
		service.NewRoomService,
		ws.NewRoomHandler,
	)
	return nil
}
