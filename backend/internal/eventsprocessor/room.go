package eventsprocessor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/incoming"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	ivalidator "github.com/holdennekt/sgame/backend/internal/interface/validator"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/internal/message"
)

const (
	OWNER_TTL          = 10 * time.Second
	OWNER_REFRESH_RATE = 5 * time.Second
	GET_URL_TTL        = 3 * time.Hour
)

type RoomEventsProcessor struct {
	client             realtime.Channel
	lobbyServer        realtime.Channel
	roomServer         realtime.Channel
	roomInternalServer realtime.Channel
	roomCache          cache.Room
	roomRepository     repository.Room
	storage            storage.Storage
	id                 string
	user               domain.User
	pack               *domain.Pack
	cfg                *config.Config
	validator          ivalidator.AnswerValidator
	isSpectator        bool
	onDisconnect       func(ctx context.Context, userId, roomId string) (*domain.Room, error)
}

type RoomEventsProcessorGetter func(client realtime.Channel, id string, user domain.User, isSpectator bool) (*RoomEventsProcessor, error)

func NewRoomEventsProcessorGetter(lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ChannelGetter, roomCache cache.Room, roomRepository repository.Room, packRepository repository.Pack, storage storage.Storage, cfg *config.Config, answerValidator ivalidator.AnswerValidator, onDisconnect func(ctx context.Context, userId, roomId string) (*domain.Room, error)) RoomEventsProcessorGetter {
	return func(client realtime.Channel, id string, user domain.User, isSpectator bool) (*RoomEventsProcessor, error) {
		room, err := roomCache.GetById(context.Background(), id)
		if err != nil {
			return nil, err
		}
		pack, err := packRepository.GetById(context.Background(), room.PackPreview.Id)
		if err != nil {
			return nil, err
		}
		return &RoomEventsProcessor{
			client:             client,
			lobbyServer:        lobbyChannelGetter.Get(domain.LOBBY),
			roomServer:         roomChannelGetter.Get(domain.ROOM_PREFIX + id),
			roomInternalServer: roomInternalChannelGetter.Get(domain.ROOM_PREFIX + id + domain.INTERNAL_POSTFIX),
			roomCache:          roomCache,
			roomRepository:     roomRepository,
			storage:            storage,
			id:                 id,
			user:               user,
			pack:               pack,
			isSpectator:        isSpectator,
			cfg:                cfg,
			validator:          answerValidator,
			onDisconnect:       onDisconnect,
		}, nil
	}
}

func (p *RoomEventsProcessor) Listen(ctx context.Context) {
	clientMessages := p.client.Receive(ctx)
	serverMessages := p.roomServer.Receive(ctx)
	for {
		select {
		case msg, ok := <-clientMessages:
			if !ok {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := p.handleClientClosure(cleanupCtx); err != nil {
					slog.Error("error while handling client closure", "err", err)
				}
				_ = p.roomServer.Close()
				return
			}

			slog.Info("user sent room message", "user", p.user.Name, "user_id", p.user.Id, "room_id", p.id, "event", msg.Event, "payload", string(msg.Payload))
			if err := p.handleClientMessage(ctx, msg); err != nil {
				slog.Error("error", "err", err)
				_ = p.client.Send(ctx, outgoing.NewErrorMessage(err))
			}
		case msg, ok := <-serverMessages:
			if !ok {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := p.handleServerClosure(cleanupCtx); err != nil {
					slog.Error("error while handling roomServer closure", "err", err)
				}
				_ = p.client.Close()
				return
			}

			slog.Info("user received room message", "user", p.user.Name, "user_id", p.user.Id, "room_id", p.id, "event", msg.Event, "payload", string(msg.Payload))
			if err := p.handleServerMessage(ctx, msg); err != nil {
				slog.Error("error", "err", err)
			}
		}
	}
}

func (p *RoomEventsProcessor) handleClientMessage(ctx context.Context, msg message.Message) error {
	if p.isSpectator {
		return nil
	}
	getURL := func(key string) (string, error) {
		return p.storage.URL(ctx, key, GET_URL_TTL)
	}
	switch msg.Event {
	case domain.Chat:
		return client.HandleClientChatMessage(ctx, p.roomServer, p.user, msg)
	case domain.StartGame:
		return incoming.HandleStartGameMessage(ctx, p.lobbyServer, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.pack, msg)
	case domain.SelectQuestion:
		return incoming.HandleSelectQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.user, p.pack, p.cfg.QuestionDemoDuration, msg)
	case domain.StartAnswer:
		return incoming.HandleStartAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.SubmitAnswer:
		return incoming.HandleSubmitAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.validator, p.cfg, msg)
	case domain.ValidateAnswer:
		return incoming.HandleValidateAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.PassQuestion:
		return incoming.HandlePassQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.PlaceBet:
		return incoming.HandlePlaceBetMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.RemoveFinalRoundCategory:
		return incoming.HandleRemoveFinalRoundCategoryMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.user, p.pack, msg)
	case domain.PlaceFinalRoundBet:
		return incoming.HandlePlaceFinalRoundBetMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.SubmitFinalRoundAnswer:
		return incoming.HandleSubmitFinalRoundAnswerMessage(ctx, p.roomServer, p.roomCache, p.id, p.user, msg)
	case domain.ValidateFinalRoundAnswer:
		return incoming.HandleValidateFinalRoundAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.user, msg)
	case domain.SkipQuestion:
		return incoming.HandleSkipQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.SkipRound:
		return incoming.HandleSkipRoundMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.user, p.pack)
	case domain.ChangeScore:
		return incoming.HandleChangeScoreMessage(ctx, p.roomServer, p.roomCache, p.id, p.user, msg)
	case domain.Pause:
		return incoming.HandlePauseMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.Unpause:
		return incoming.HandleUnpauseMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.BanPlayer:
		return incoming.HandleBanPlayerMessage(ctx, p.roomServer, p.roomCache, p.id, p.user, msg)
	}
	return nil
}

func (p *RoomEventsProcessor) handleClientClosure(ctx context.Context) error {
	slog.Info("room client channel closed", "user", p.user.Name, "user_id", p.user.Id, "room_id", p.id)

	if _, err := p.onDisconnect(ctx, p.user.Id, p.id); err != nil {
		return err
	}

	if p.isSpectator {
		return nil
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(p.id)
	if err := p.roomServer.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}
	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has disconnected", p.user.Name))
	if err := p.roomServer.Send(ctx, chatMessage); err != nil {
		return err
	}
	userDisconnected := server.NewUserDisconnectedMessage(p.id)
	return p.roomInternalServer.Send(ctx, userDisconnected)
}

func (p *RoomEventsProcessor) handleServerMessage(ctx context.Context, msg message.Message) error {
	switch msg.Event {
	case domain.Chat:
		return client.HandleServerChatMessage(ctx, p.client, msg)
	case domain.RoomUpdated:
		return outgoing.HandleRoomUpdatedMessage(ctx, p.roomCache, p.client, p.user, msg)
	case domain.RoundDemo:
		return outgoing.HandleRoundDemoMessage(ctx, p.client, msg)
	case domain.QuestionDemo:
		return outgoing.HandleQuestionDemoMessage(ctx, p.client, msg)
	case domain.CorrectAnswerDemo:
		return outgoing.HandleCorrectAnswerDemoMessage(ctx, p.client, msg)
	case domain.RoomDeleted:
		_ = p.lobbyServer.Close()
		_ = p.roomServer.Close()
		return outgoing.HandleRoomDeletedMessage(ctx, p.client, msg)
	}
	return nil
}

func (p *RoomEventsProcessor) handleServerClosure(_ context.Context) error {
	slog.Info("room server channel closed", "user", p.user.Name, "user_id", p.user.Id, "room_id", p.id)
	return nil
}

type RoomInternalEventsProcessor struct {
	lobbyServer        realtime.Channel
	roomServer         realtime.Channel
	roomInternalServer realtime.Channel
	roomCache          cache.Room
	roomRepository     repository.Room
	storage            storage.Storage
	id                 string
	pack               *domain.Pack
	cfg                *config.Config
}

type RoomInternalEventsProcessorGetter func(id string) (*RoomInternalEventsProcessor, error)

func NewRoomInternalEventsProcessorGetter(lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ChannelGetter, roomCache cache.Room, roomRepository repository.Room, packRepository repository.Pack, storage storage.Storage, cfg *config.Config) RoomInternalEventsProcessorGetter {
	return func(id string) (*RoomInternalEventsProcessor, error) {
		room, err := roomCache.GetById(context.Background(), id)
		if err != nil {
			return nil, err
		}
		pack, err := packRepository.GetById(context.Background(), room.PackPreview.Id)
		if err != nil {
			return nil, err
		}
		return &RoomInternalEventsProcessor{
			lobbyServer:        lobbyChannelGetter.Get(domain.LOBBY),
			roomServer:         roomChannelGetter.Get(domain.ROOM_PREFIX + id),
			roomInternalServer: roomInternalChannelGetter.Get(domain.ROOM_PREFIX + id + domain.INTERNAL_POSTFIX),
			roomCache:          roomCache,
			roomRepository:     roomRepository,
			storage:            storage,
			id:                 id,
			pack:               pack,
			cfg:                cfg,
		}, nil
	}
}

func (p *RoomInternalEventsProcessor) Listen(ctx context.Context) {
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.handleServerClosure(cleanupCtx); err != nil {
			slog.Error("error while handling internalRoomServer closure", "err", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		tickerC := time.Tick(OWNER_REFRESH_RATE)
		for {
			select {
			case <-ctx.Done():
				slog.Debug("exiting owner refresh ticker")
				return
			case <-tickerC:
				_ = p.roomCache.UpdateOwner(ctx, p.id, OWNER_TTL)
			}
		}
	}()

	messages := p.roomInternalServer.Receive(ctx)
	for {
		msg, ok := <-messages
		if !ok {
			slog.Info("internal room server channel was closed")
			return
		}

		slog.Info("server received room internal message", "room_id", p.id, "event", msg.Event, "payload", string(msg.Payload))
		if err := p.handleMessage(ctx, msg); err != nil {
			slog.Error("error", "err", err)
			return
		}
	}
}

func (p *RoomInternalEventsProcessor) handleMessage(ctx context.Context, msg message.Message) error {
	getURL := func(key string) (string, error) {
		return p.storage.URL(ctx, key, GET_URL_TTL)
	}
	switch msg.Event {
	case domain.RoundStarted:
		return server.HandleRoundStartedMessage(ctx, p.roomServer, p.roomCache, p.id, p.pack)
	case domain.RevealingStarted:
		return server.HandleRevealingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, msg)
	case domain.QuestionStarted:
		return server.HandleQuestionStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, msg)
	case domain.AnswerStarted:
		return server.HandleAnswerStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, msg)
	case domain.QuestionEnded:
		return server.HandleQuestionEndedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.pack, msg)
	case domain.PassingStarted:
		return server.HandlePassingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, msg)
	case domain.BettingStarted:
		return server.HandleBettingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, msg)
	case domain.FinalRoundBettingStarted:
		return server.HandleFinalRoundBettingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
	case domain.FinalRoundQuestionStarted:
		return server.HandleFinalRoundQuestionStartedMessage(ctx, p.roomServer, p.roomCache, p.id)
	case domain.GameEnded:
		return server.HandleGameEndedMessage(ctx, p.roomServer, p.roomInternalServer, p.lobbyServer, p.roomCache, p.roomRepository, p.id, time.Duration(p.cfg.IdleRoomTTL)*time.Second)
	case domain.UserDisconnected:
		return server.HandleUserDisconnectedMessage(ctx, p.roomServer, p.roomInternalServer, p.lobbyServer, p.roomCache, p.roomRepository, p.id, msg, time.Duration(p.cfg.IdleRoomTTL)*time.Second)
	case domain.RoomDeleted:
		slog.Info("internal room server got room_deleted event")
		_ = p.lobbyServer.Close()
		_ = p.roomServer.Close()
		return p.roomInternalServer.Close()
	}
	return nil
}

func (p *RoomInternalEventsProcessor) handleServerClosure(ctx context.Context) error {
	slog.Info("room channel closed", "room_id", p.id)
	if err := p.roomServer.Delete(ctx); err != nil {
		return err
	}
	return p.roomInternalServer.Delete(ctx)
}
