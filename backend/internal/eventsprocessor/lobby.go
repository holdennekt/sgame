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

func NewLobbyEventsProcessorGetter(lobbyChannelGetter realtime.ChannelGetter, roomCache cache.Room) LobbyEventsProcessorGetter {
	return func(client realtime.Channel, user domain.User) *LobbyEventsProcessor {
		return &LobbyEventsProcessor{client, lobbyChannelGetter.Get(domain.LOBBY), roomCache, user}
	}
}

func (p *LobbyEventsProcessor) Listen(ctx context.Context) {
	clientMessages := p.client.Receive(ctx)
	serverMessages := p.server.Receive(ctx)
	for {
		select {
		case msg, ok := <-clientMessages:
			if !ok {
				if err := p.handleClientClosure(); err != nil {
					slog.Error("error while handling client closure", "err", err)
				}
				_ = p.server.Close()
				return
			}

			slog.Info("user sent lobby message", "user", p.user.Name, "user_id", p.user.Id, "event", msg.Event, "payload", string(msg.Payload))
			if err := p.handleClientMessage(ctx, msg); err != nil {
				slog.Error("error", "err", err)
				_ = p.client.Send(ctx, outgoing.NewErrorMessage(err))
			}

		case msg, ok := <-serverMessages:
			if !ok {
				if err := p.handleServerClosure(); err != nil {
					slog.Error("error while handling server closure", "err", err)
				}
				_ = p.client.Close()
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
		if err := client.HandleClientChatMessage(ctx, p.server, p.user, msg); err != nil {
			slog.Error("error handling client chat message", "err", err)
		}
	}
	return nil
}

func (p *LobbyEventsProcessor) handleServerMessage(ctx context.Context, msg message.Message) error {
	switch msg.Event {
	case domain.Chat:
		if err := client.HandleServerChatMessage(ctx, p.client, msg); err != nil {
			slog.Error("error handling server chat message", "err", err)
		}
	case domain.RoomUpdated:
		if err := outgoing.HandleLobbyRoomUpdatedMessage(ctx, p.roomCache, p.client, msg); err != nil {
			slog.Error("error handling room updated message", "err", err)
		}
	case domain.RoomDeleted:
		if err := outgoing.HandleRoomDeletedMessage(ctx, p.client, msg); err != nil {
			slog.Error("error handling room deleted message", "err", err)
		}
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
