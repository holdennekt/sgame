//go:build wireinject
// +build wireinject

package app

import (
	"context"

	"github.com/google/wire"
	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	redisCache "github.com/holdennekt/sgame/backend/internal/infrastructure/cache/redis"
	mongoDatabase "github.com/holdennekt/sgame/backend/internal/infrastructure/database/mongo"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/pubsub"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/streams"
	infravalidator "github.com/holdennekt/sgame/backend/internal/infrastructure/validator"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
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
	realtime.ChannelGetter
}

type StreamsChannelGetter struct {
	realtime.ChannelGetter
}

type StreamsPersistentChannelGetter struct {
	realtime.ChannelGetter
}

func provideManager(client *redis.Client) *pubsub.Manager {
	return pubsub.NewManager(client)
}

func providePubSubChannelGetter(client *redis.Client, manager *pubsub.Manager) PubSubChannelGetter {
	return PubSubChannelGetter{pubsub.NewManagedChannelGetter(client, manager)}
}

func provideStreamsChannelGetter(client *redis.Client, manager *pubsub.Manager) StreamsChannelGetter {
	return StreamsChannelGetter{streams.NewManagedChannelGetter(client, manager)}
}

func provideStreamsPersistentChannelGetter(client *redis.Client, manager *pubsub.Manager) StreamsPersistentChannelGetter {
	return StreamsPersistentChannelGetter{streams.NewPersistentManagedChannelGetter(client, manager)}
}

func provideLobbyEventsProcessorGetter(roomCache cache.Room, pubsubGetter PubSubChannelGetter) eventsprocessor.LobbyEventsProcessorGetter {
	return eventsprocessor.NewLobbyEventsProcessorGetter(pubsubGetter.ChannelGetter, roomCache)
}

func provideAnswerValidator(cfg *config.Config) ivalidator.AnswerValidator {
	switch cfg.ValidatorType {
	case "ollama":
		return infravalidator.NewLocalLLMValidator(cfg.OllamaURL, cfg.OllamaSystemPrompt)
	case "gemini":
		v, err := infravalidator.NewGeminiValidator(context.Background(), cfg.GeminiProjectID, cfg.GeminiLocation, cfg.GeminiSystemPrompt)
		if err != nil {
			panic("failed to initialize Gemini validator: " + err.Error())
		}
		return v
	default:
		return nil
	}
}

func provideRoomEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentGetter StreamsPersistentChannelGetter, cfg *config.Config, validator ivalidator.AnswerValidator, roomService *service.RoomService) eventsprocessor.RoomEventsProcessorGetter {
	return eventsprocessor.NewRoomEventsProcessorGetter(pubsubGetter.ChannelGetter, streamsGetter.ChannelGetter, persistentGetter.ChannelGetter, roomCache, roomRepo, packRepo, storage, cfg, validator, roomService.Disconnect)
}

func provideRoomInternalEventsProcessorGetter(roomCache cache.Room, roomRepo repository.Room, packRepo repository.Pack, storage storage.Storage, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentGetter StreamsPersistentChannelGetter, cfg *config.Config) eventsprocessor.RoomInternalEventsProcessorGetter {
	return eventsprocessor.NewRoomInternalEventsProcessorGetter(pubsubGetter.ChannelGetter, streamsGetter.ChannelGetter, persistentGetter.ChannelGetter, roomCache, roomRepo, packRepo, storage, cfg)
}

func provideRoomService(packRepository repository.Pack, roomRepository repository.Room, roomCache cache.Room, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentGetter StreamsPersistentChannelGetter, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter, cfg *config.Config, validator ivalidator.AnswerValidator) *service.RoomService {
	return service.NewRoomService(packRepository, roomRepository, roomCache, pubsubGetter.ChannelGetter, streamsGetter.ChannelGetter, persistentGetter.ChannelGetter, roomInternalEventsProcessorGetter, cfg, validator)
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
	return ws.NewLobbyHandler(pubsubGetter.ChannelGetter, lobbyEventsProcessorGetter)
}

func provideRoomHandler(roomService *service.RoomService, roomEventsProcessorGetter eventsprocessor.RoomEventsProcessorGetter, pubsubGetter PubSubChannelGetter, streamsGetter StreamsChannelGetter, persistentGetter StreamsPersistentChannelGetter) *ws.RoomHandler {
	return ws.NewRoomHandler(roomService, pubsubGetter.ChannelGetter, streamsGetter.ChannelGetter, persistentGetter.ChannelGetter, roomEventsProcessorGetter)
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
		provideManager,
		providePubSubChannelGetter,
		provideStreamsChannelGetter,
		provideStreamsPersistentChannelGetter,
		provideLobbyEventsProcessorGetter,
		provideRoomEventsProcessorGetter,
		provideRoomInternalEventsProcessorGetter,
		provideAnswerValidator,
		NewApp,
	)
	return nil
}
