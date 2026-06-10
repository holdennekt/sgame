package eventsprocessor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type LobbyEventsProcessor struct {
	client    realtime.Channel
	server    realtime.Channel
	roomCache cache.Room
	user      domain.User
}

type LobbyEventsProcessorGetter func(client realtime.Channel, user domain.User) *LobbyEventsProcessor

func NewLobbyEventsProcessorGetter(lobbyChannelGetter realtime.ServerChannelGetter, roomCache cache.Room) LobbyEventsProcessorGetter {
	return func(client realtime.Channel, user domain.User) *LobbyEventsProcessor {
		return &LobbyEventsProcessor{client, lobbyChannelGetter.Get(domain.LOBBY), roomCache, user}
	}
}

func (p *LobbyEventsProcessor) Listen(ctx context.Context) {
	clientMessages := p.client.Recieve(ctx)
	serverMessages := p.server.Recieve(ctx)
	for {
		select {
		case msg, ok := <-clientMessages:
			if !ok {
				if err := p.handleClientClosure(); err != nil {
					slog.Error("error while handling client closure", "err", err)
				}
				p.server.Close()
				return
			}

			slog.Info("user sent lobby message", "user", p.user.Name, "user_id", p.user.Id, "event", msg.Event, "payload", string(msg.Payload))
			if err := p.handleClientMessage(ctx, msg); err != nil {
				slog.Error("error", "err", err)
				p.client.Send(ctx, outgoing.NewErrorMessage(err))
			}

		case msg, ok := <-serverMessages:
			if !ok {
				if err := p.handleServerClosure(); err != nil {
					slog.Error("error while handling server closure", "err", err)
				}
				p.client.Close()
				return
			}

			slog.Info("user received lobby message", "user", p.user.Name, "user_id", p.user.Id, "event", msg.Event, "payload", string(msg.Payload))
			if err := p.handleServerMessage(ctx, msg); err != nil {
				slog.Error("error", "err", err)
			}
		}
	}
}

func (p *LobbyEventsProcessor) handleClientMessage(ctx context.Context, msg message.Message) error {
	switch msg.Event {
	case domain.Chat:
		client.HandleClientChatMessage(ctx, p.server, p.user, msg)
	}
	return nil
}

func (p *LobbyEventsProcessor) handleServerMessage(ctx context.Context, msg message.Message) error {
	switch msg.Event {
	case domain.Chat:
		client.HandleServerChatMessage(ctx, p.client, msg)
	case domain.RoomUpdated:
		outgoing.HandleRoomUpdatedMessage(ctx, p.roomCache, p.client, p.user, msg)
	case domain.RoomDeleted:
		outgoing.HandleRoomDeletedMessage(ctx, p.client, msg)
	}
	return nil
}

func (p *LobbyEventsProcessor) handleClientClosure() error {
	slog.Info("lobby client channel closed", "user", p.user.Name, "user_id", p.user.Id)
	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has disconnected", p.user.Name))
	return p.server.Send(context.Background(), chatMessage)
}

func (p *LobbyEventsProcessor) handleServerClosure() error {
	slog.Info("lobby server channel closed", "user", p.user.Name, "user_id", p.user.Id)
	return nil
}
