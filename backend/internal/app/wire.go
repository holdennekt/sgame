//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/holdennekt/sgame/backend/internal/config"
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
	mongoDatabase.NewPackDraftRepository,
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

type PersistentStreamsChannelGetter struct {
	realtime.ServerChannelGetter
}

func providePubSubChannelGetter(client *redis.Client) PubSubChannelGetter {
	return PubSubChannelGetter{pubsub.NewManagedServerChannelGetter(client, pubsub.NewManager(client))}
}

func provideStreamsChannelGetter(client *redis.Client, manager *streams.StreamManager) StreamsChannelGetter {
	return StreamsChannelGetter{streams.NewManagedServerChannelGetter(client, manager, false)}
}

func providePersistentStreamsChannelGetter(client *redis.Client, manager *streams.StreamManager) PersistentStreamsChannelGetter {
	return PersistentStreamsChannelGetter{streams.NewManagedServerChannelGetter(client, manager, true)}
}

func provideLobbyEventsProcessorGetter(roomCache cache.Room, pubsubGetter PubSubChannelGetter) eventsprocessor.LobbyEventsProcessorGetter {
	return eventsprocessor.NewLobbyEventsProcessorGetter(pubsubGetter.ServerChannelGetter, roomCache)
}

func provideRoomEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentStreamsGetter PersistentStreamsChannelGetter, cfg *config.Config) eventsprocessor.RoomEventsProcessorGetter {
	return eventsprocessor.NewRoomEventsProcessorGetter(pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, persistentStreamsGetter.ServerChannelGetter, roomCache, roomRepo, packRepo, storage, cfg)
}

func provideRoomInternalEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentStreamsGetter PersistentStreamsChannelGetter, cfg *config.Config) eventsprocessor.RoomInternalEventsProcessorGetter {
	return eventsprocessor.NewRoomInternalEventsProcessorGetter(pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, persistentStreamsGetter.ServerChannelGetter, roomCache, roomRepo, packRepo, storage, cfg)
}

func provideRoomService(packRepository repository.Pack, roomRepository repository.Room, roomCache cache.Room, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentStreamsGetter PersistentStreamsChannelGetter, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter, cfg *config.Config) *service.RoomService {
	return service.NewRoomService(packRepository, roomRepository, roomCache, pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, persistentStreamsGetter.ServerChannelGetter, roomInternalEventsProcessorGetter, cfg)
}

var ServiceSet = wire.NewSet(
	service.NewAuthService,
	service.NewUserService,
	provideRoomService,
	service.NewAttachmentService,
	service.NewPackService,
	service.NewPackDraftService,
)

var ControllerSet = wire.NewSet(
	http.NewAuthController,
	http.NewUserController,
	http.NewPackController,
	http.NewPackDraftController,
	http.NewRoomController,
)

func provideLobbyHandler(pubsubGetter PubSubChannelGetter, lobbyEventsProcessorGetter eventsprocessor.LobbyEventsProcessorGetter) *ws.LobbyHandler {
	return ws.NewLobbyHandler(pubsubGetter.ServerChannelGetter, lobbyEventsProcessorGetter)
}

func provideRoomHandler(roomService *service.RoomService, roomEventsProcessorGetter eventsprocessor.RoomEventsProcessorGetter, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentStreamsGetter PersistentStreamsChannelGetter) *ws.RoomHandler {
	return ws.NewRoomHandler(roomService, pubsubGetter.ServerChannelGetter, streamsGetter.ServerChannelGetter, persistentStreamsGetter.ServerChannelGetter, roomEventsProcessorGetter)
}

var HandlerSet = wire.NewSet(
	provideLobbyHandler,
	provideRoomHandler,
)

func InitializeApp(mdb *mongo.Database, rds *redis.Client, storage storage.Storage, cfg *config.Config) *app {
	wire.Build(
		RepoSet,
		CacheSet,
		ServiceSet,
		ControllerSet,
		HandlerSet,
		streams.NewStreamManager,
		providePubSubChannelGetter,
		provideStreamsChannelGetter,
		providePersistentStreamsChannelGetter,
		provideLobbyEventsProcessorGetter,
		provideRoomEventsProcessorGetter,
		provideRoomInternalEventsProcessorGetter,
		NewApp,
	)
	return nil
}
