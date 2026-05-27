//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	redisCache "github.com/holdennekt/sgame/backend/internal/infrastructure/cache/redis"
	mongoDatabase "github.com/holdennekt/sgame/backend/internal/infrastructure/database/mongo"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/pubsub"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/streams"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/internal/transport/http"
	"github.com/holdennekt/sgame/backend/internal/transport/ws"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var RepoSet = wire.NewSet(
	mongoDatabase.NewUserRepository,
	mongoDatabase.NewRoomRepository,
	mongoDatabase.NewPackRepository,
)

var CacheSet = wire.NewSet(
	redisCache.NewSessionCache,
	redisCache.NewRoomCache,
)

type PubSubChannelGetter struct {
	realtime.ServerChannelGetter
}

type StreamsChannelGetter struct {
	realtime.ServerChannelGetter
}

func providePubSubChannelGetter(client *redis.Client) PubSubChannelGetter {
	return PubSubChannelGetter{pubsub.NewServerChannelGetter(client)}
}

func provideStreamsChannelGetter(client *redis.Client) StreamsChannelGetter {
	return StreamsChannelGetter{streams.NewServerChannelGetter(client)}
}

func provideLobbyEventsProcessorGetter(roomCache cache.Room, pubsubGetter PubSubChannelGetter) eventsprocessor.LobbyEventsProcessorGetter {
	return eventsprocessor.NewLobbyEventsProcessorGetter(pubsubGetter.ServerChannelGetter, roomCache)
}

func provideRoomEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter) eventsprocessor.RoomEventsProcessorGetter {
	return eventsprocessor.NewRoomEventsProcessorGetter(pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, roomCache, roomRepo, packRepo, storage)
}

func provideRoomInternalEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter) eventsprocessor.RoomInternalEventsProcessorGetter {
	return eventsprocessor.NewRoomInternalEventsProcessorGetter(pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, roomCache, roomRepo, packRepo, storage)
}

func provideRoomService(userRepository repository.User, packRepository repository.Pack, roomRepository repository.Room, roomCache cache.Room, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter) *service.RoomService {
	return service.NewRoomService(userRepository, packRepository, roomRepository, roomCache, pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, roomInternalEventsProcessorGetter)
}

var ServiceSet = wire.NewSet(
	service.NewAuthService,
	service.NewUserService,
	provideRoomService,
	service.NewPackService,
)

var ControllerSet = wire.NewSet(
	http.NewAuthController,
	http.NewUserController,
	http.NewPackController,
	http.NewRoomController,
)

func provideLobbyHandler(userService *service.UserService, pubsubGetter PubSubChannelGetter, lobbyEventsProcessorGetter eventsprocessor.LobbyEventsProcessorGetter) *ws.LobbyHandler {
	return ws.NewLobbyHandler(userService, pubsubGetter.ServerChannelGetter, lobbyEventsProcessorGetter)
}

func provideRoomHandler(roomService *service.RoomService, userService *service.UserService, roomEventsProcessorGetter eventsprocessor.RoomEventsProcessorGetter, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter) *ws.RoomHandler {
	return ws.NewRoomHandler(roomService, userService, pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, roomEventsProcessorGetter)
}

var HandlerSet = wire.NewSet(
	provideLobbyHandler,
	provideRoomHandler,
)

func InitializeApp(mdb *mongo.Database, rds *redis.Client, storage storage.Storage) *app {
	wire.Build(
		RepoSet,
		CacheSet,
		ServiceSet,
		ControllerSet,
		HandlerSet,
		providePubSubChannelGetter,
		provideStreamsChannelGetter,
		provideLobbyEventsProcessorGetter,
		provideRoomEventsProcessorGetter,
		provideRoomInternalEventsProcessorGetter,
		NewApp,
	)
	return nil
}
