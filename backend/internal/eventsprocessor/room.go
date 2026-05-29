package eventsprocessor

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/incoming"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
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
}

type RoomEventsProcessorGetter func(client realtime.Channel, id string, user domain.User) (*RoomEventsProcessor, error)

func NewRoomEventsProcessorGetter(lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, roomRepository repository.Room, packRepository repository.Pack, storage storage.Storage) RoomEventsProcessorGetter {
	return func(client realtime.Channel, id string, user domain.User) (*RoomEventsProcessor, error) {
		room, err := roomCache.GetById(context.Background(), id)
		if err != nil {
			return nil, err
		}
		pack, err := packRepository.GetById(context.Background(), room.PackPreview.Id)
		if err != nil {
			return nil, err
		}
		return &RoomEventsProcessor{
			client,
			lobbyChannelGetter.Get(domain.LOBBY),
			roomChannelGetter.Get(domain.ROOM_PREFIX + id),
			roomInternalChannelGetter.Get(domain.ROOM_PREFIX + id + domain.INTERNAL_POSTFIX),
			roomCache,
			roomRepository,
			storage,
			id,
			user,
			pack,
		}, nil
	}
}

func (p *RoomEventsProcessor) Listen(ctx context.Context) {
	clientMessages := p.client.Recieve(ctx)
	serverMessages := p.roomServer.Recieve(ctx)
	for {
		select {
		case msg, ok := <-clientMessages:
			if !ok {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := p.handleClientClosure(ctx); err != nil {
					log.Println("error while handling client closure", err)
				}
				p.roomServer.Close()
				return
			}

			log.Printf("User \"%s(%s)\" has sent room \"%s\" message with event \"%s\": %v\n", p.user.Name, p.user.Id, p.id, msg.Event, string(msg.Payload))
			if err := p.handleClientMessage(ctx, msg); err != nil {
				log.Println(err)
				p.client.Send(ctx, outgoing.NewErrorMessage(err))
			}
		case msg, ok := <-serverMessages:
			if !ok {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := p.handleServerClosure(ctx); err != nil {
					log.Println("error while handling roomServer closure", err)
				}
				p.client.Close()
				return
			}

			log.Printf("User \"%s(%s)\" has recieved room \"%s\" message with event \"%s\": %v\n", p.user.Name, p.user.Id, p.id, msg.Event, string(msg.Payload))
			if err := p.handleServerMessage(ctx, msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func (p *RoomEventsProcessor) handleClientMessage(ctx context.Context, msg message.Message) error {
	getURL := func(key string) (string, error) {
		return p.storage.URL(ctx, key, GET_URL_TTL)
	}
	switch msg.Event {
	case domain.Chat:
		return client.HandleClientChatMessage(ctx, p.roomServer, p.user, msg)
	case domain.StartGame:
		return incoming.HandleStartGameMessage(ctx, p.lobbyServer, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.pack, msg)
	case domain.SelectQuestion:
		return incoming.HandleSelectQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, getURL, p.id, p.user, p.pack, msg)
	case domain.SubmitAnswer:
		return incoming.HandleSubmitAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
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
		return incoming.HandleValidateFinalRoundAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.SkipQuestion:
		return incoming.HandleSkipQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.ChangeScore:
		return incoming.HandleChangeScoreMessage(ctx, p.roomServer, p.roomCache, p.id, p.user, msg)
	case domain.Pause:
		return incoming.HandlePauseMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.Unpause:
		return incoming.HandleUnpauseMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	}
	return nil
}

func (p *RoomEventsProcessor) handleClientClosure(ctx context.Context) error {
	log.Printf("User \"%s(%s)\" room \"%s\" client channel closed\n", p.user.Name, p.user.Id, p.id)
	_, err := p.roomCache.SafeSet(ctx, p.id, func(room *domain.Room) error {
		playerIndex := slices.IndexFunc(room.Players, func(player domain.Player) bool {
			return p.user.Id == player.Id
		})
		userInRoom := room.IsUserHost(p.user.Id) || playerIndex != -1
		if !userInRoom {
			return nil
		}

		if room.IsUserHost(p.user.Id) {
			room.Host.IsConnected = false
		} else {
			room.Players[playerIndex].IsConnected = false
		}
		return nil
	})
	if err != nil {
		return err
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
		return outgoing.HandleRoomDeletedMessage(ctx, p.client, msg)
	}
	return nil
}

func (p *RoomEventsProcessor) handleServerClosure(ctx context.Context) error {
	log.Printf("User \"%s(%s)\" room \"%s\" roomServer channel closed\n", p.user.Name, p.user.Id, p.id)
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
}

type RoomInternalEventsProcessorGetter func(id string) (*RoomInternalEventsProcessor, error)

func NewRoomInternalEventsProcessorGetter(lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, roomRepository repository.Room, packRepository repository.Pack, storage storage.Storage) RoomInternalEventsProcessorGetter {
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
			lobbyChannelGetter.Get(domain.LOBBY),
			roomChannelGetter.Get(domain.ROOM_PREFIX + id),
			roomInternalChannelGetter.Get(domain.ROOM_PREFIX + id + domain.INTERNAL_POSTFIX),
			roomCache,
			roomRepository,
			storage,
			id,
			pack,
		}, nil
	}
}

func (p *RoomInternalEventsProcessor) Listen(ctx context.Context) error {
	defer p.handleServerClosure(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		tickerC := time.Tick(OWNER_REFRESH_RATE)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerC:
				p.roomCache.UpdateOwner(ctx, p.id, OWNER_TTL)
			}
		}
	}()

	messages := p.roomInternalServer.Recieve(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messages:
			log.Printf("Server has recieved room \"%s\" internal message with event \"%s\": %v\n", p.id, msg.Event, string(msg.Payload))
			if err := p.handleMessage(ctx, msg); err != nil {
				log.Println(err)
				return err
			}
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
		return server.HandleGameEndedMessage(ctx, p.roomServer, p.lobbyServer, p.roomCache, p.roomRepository, p.id)
	case domain.UserDisconnected:
		return server.HandleUserDisconnectedMessage(ctx, p.roomServer, p.lobbyServer, p.roomCache, p.roomRepository, p.id, msg)
	case domain.RoomDeleted:
		return p.roomServer.Close()
	}
	return nil
}

func (p *RoomInternalEventsProcessor) handleServerClosure(ctx context.Context) error {
	log.Printf("Server room \"%s\" channel closed\n", p.id)
	return p.roomInternalServer.Close()
}
