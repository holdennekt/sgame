package eventsprocessor

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/incoming"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/message"
)

type RoomEventsProcessor struct {
	client             realtime.Channel
	lobbyServer        realtime.Channel
	roomServer         realtime.Channel
	roomInternalServer realtime.Channel
	roomCache          cache.Room
	roomRepository     repository.Room
	id                 string
	user               domain.User
	pack               *domain.Pack
}

func NewRoomEventsProcessor(client realtime.Channel, lobbyServer realtime.Channel, roomServer realtime.Channel, roomInternalServer realtime.Channel, roomCache cache.Room, roomRepository repository.Room, id string, user domain.User, pack *domain.Pack) *RoomEventsProcessor {
	return &RoomEventsProcessor{client, lobbyServer, roomServer, roomInternalServer, roomCache, roomRepository, id, user, pack}
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
	switch msg.Event {
	case domain.Chat:
		return client.HandleClientChatMessage(ctx, p.roomServer, p.user, msg)
	case domain.StartGame:
		return incoming.HandleStartGameMessage(ctx, p.lobbyServer, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.pack, msg)
	case domain.SelectQuestion:
		return incoming.HandleSelectQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.pack, msg)
	case domain.SubmitAnswer:
		return incoming.HandleSubmitAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.ValidateAnswer:
		return incoming.HandleValidateAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.PassQuestion:
		return incoming.HandlePassQuestionMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.PlaceBet:
		return incoming.HandlePlaceBetMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.RemoveFinalRoundCategory:
		return incoming.HandleRemoveFinalRoundCategoryMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, p.pack, msg)
	case domain.PlaceFinalRoundBet:
		return incoming.HandlePlaceFinalRoundBetMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
	case domain.SubmitFinalRoundAnswer:
		return incoming.HandleSubmitFinalRoundAnswerMessage(ctx, p.roomServer, p.roomCache, p.id, p.user, msg)
	case domain.ValidateFinalRoundAnswer:
		return incoming.HandleValidateFinalRoundAnswerMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.user, msg)
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

func (p *RoomEventsProcessor) ListenInternal(ctx context.Context) error {
	for msg := range p.roomInternalServer.Recieve(ctx) {
		log.Printf("Server has recieved room \"%s\" internal message with event \"%s\": %v\n", p.id, msg.Event, string(msg.Payload))
		if err := p.handleInternalMessage(ctx, msg); err != nil {
			log.Println(err)
			return err
		}
	}
	p.handleInternalServerClosure(ctx)
	return nil
}

func (p *RoomEventsProcessor) handleInternalMessage(ctx context.Context, msg message.Message) error {
	switch msg.Event {
	case domain.RoundStarted:
		return server.HandleRoundStartedMessage(ctx, p.roomServer, p.roomCache, p.id, p.pack)
	case domain.RevealingStarted:
		return server.HandleRevealingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
	case domain.QuestionStarted:
		return server.HandleQuestionStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
	case domain.AnswerStarted:
		return server.HandleAnswerStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
	case domain.QuestionEnded:
		return server.HandleQuestionEndedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id, p.pack, msg)
	case domain.PassingStarted:
		return server.HandlePassingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
	case domain.BettingStarted:
		return server.HandleBettingStartedMessage(ctx, p.roomServer, p.roomInternalServer, p.roomCache, p.id)
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

func (p *RoomEventsProcessor) handleInternalServerClosure(ctx context.Context) error {
	log.Printf("Server room \"%s\" channel closed\n", p.id)
	return nil
}
