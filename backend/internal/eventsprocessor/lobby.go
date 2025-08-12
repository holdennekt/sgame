package eventsprocessor

import (
	"context"
	"fmt"
	"log"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/message"
)

type LobbyEventsProcessor struct {
	client    realtime.Channel
	server    realtime.Channel
	roomCache cache.Room
	user      domain.User
}

func NewLobbyEventsProcessor(client realtime.Channel, server realtime.Channel, roomCache cache.Room, user domain.User) *LobbyEventsProcessor {
	return &LobbyEventsProcessor{client, server, roomCache, user}
}

func (p *LobbyEventsProcessor) Listen(ctx context.Context) {
	clientMessages := p.client.Recieve(ctx)
	serverMessages := p.server.Recieve(ctx)
	for {
		select {
		case msg, ok := <-clientMessages:
			if !ok {
				if err := p.handleClientClosure(); err != nil {
					log.Println("error while handling client closure", err)
				}
				p.server.Close()
				return
			}

			log.Printf("User \"%s(%s)\" has sent lobby message with event \"%s\": %v\n", p.user.Name, p.user.Id, msg.Event, string(msg.Payload))
			if err := p.handleClientMessage(ctx, msg); err != nil {
				log.Println(err)
				p.client.Send(ctx, outgoing.NewErrorMessage(err))
			}

		case msg, ok := <-serverMessages:
			if !ok {
				if err := p.handleServerClosure(); err != nil {
					log.Println("error while handling server closure", err)
				}
				p.client.Close()
				return
			}

			log.Printf("User \"%s(%s)\" has recieved lobby message with event \"%s\": %v\n", p.user.Name, p.user.Id, msg.Event, string(msg.Payload))
			if err := p.handleServerMessage(ctx, msg); err != nil {
				log.Println(err)
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
	log.Printf("User \"%s(%s)\" lobby client channel closed\n", p.user.Name, p.user.Id)
	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has disconnected", p.user.Name))
	return p.server.Send(context.Background(), chatMessage)
}

func (p *LobbyEventsProcessor) handleServerClosure() error {
	log.Printf("User \"%s(%s)\" lobby server channel closed\n", p.user.Name, p.user.Id)
	return nil
}
