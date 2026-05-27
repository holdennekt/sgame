package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

const UPDATE_ROOM_RETRIES = 3

type RoomService struct {
	packRepository                    repository.Pack
	roomRepository                    repository.Room
	roomCache                         cache.Room
	lobbyChannelGetter                realtime.ServerChannelGetter
	roomChannelGetter                 realtime.ServerChannelGetter
	roomInternalChannelGetter         realtime.ServerChannelGetter
	roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter
}

func NewRoomService(packRepository repository.Pack, roomRepository repository.Room, roomCache cache.Room, lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ServerChannelGetter, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter) *RoomService {
	return &RoomService{packRepository, roomRepository, roomCache, lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter, roomInternalEventsProcessorGetter}
}

func (s *RoomService) Create(ctx context.Context, userId string, crr dto.CreateRoomRequest) (string, error) {
	pack, err := s.packRepository.GetById(ctx, crr.PackId)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}

	room := &domain.Room{
		Id:      id.String(),
		Name:    crr.Name,
		Options: crr.Options,
		PackPreview: domain.PackPreview{
			Id:   pack.Id,
			Name: pack.Name,
		},
		CreatedBy: userId,
		Players:   make([]domain.Player, 0),
		State:     domain.WaitingForStart,
	}

	if err := s.roomCache.Set(ctx, room); err != nil {
		return "", custerr.NewInternalErr(err)
	}

	_, err = s.roomCache.TrySetOwner(ctx, room.Id, eventsprocessor.OWNER_TTL)
	if err != nil {
		return "", err
	}
	processor, err := s.roomInternalEventsProcessorGetter(room.Id)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	go processor.Listen(context.Background())

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(room.Id)
	lobbyServerChannel := s.lobbyChannelGetter.Get(domain.LOBBY)
	if err := lobbyServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return "", err
	}
	return room.Id, nil
}

func (s *RoomService) GetById(ctx context.Context, id string) (*domain.Room, error) {
	return s.roomCache.GetById(ctx, id)
}

func (s *RoomService) GetProjection(ctx context.Context, userId, id, password string) (any, error) {
	room, err := s.roomCache.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if room.Options.Type == domain.Private && *room.Options.Password != password {
		return nil, custerr.NewForbiddenErr("wrong password")
	}

	return room.GetProjection(userId), nil
}

func (s *RoomService) Get(ctx context.Context) ([]domain.RoomLobby, error) {
	return s.roomCache.Get(ctx)
}

func (s *RoomService) Join(ctx context.Context, user domain.User, id, password string) (any, error) {
	newRoom, err := s.roomCache.SafeSet(ctx, id, func(room *domain.Room) error {
		if room.IsUserIn(user.Id) {
			return nil
		}
		if room.IsUserBanned(user.Id) {
			return custerr.NewForbiddenErr("you were banned from this room")
		}
		if room.Options.Type == domain.Private && *room.Options.Password != password {
			return custerr.NewForbiddenErr("wrong password")
		}

		full := len(room.Players) >= room.Options.MaxPlayers
		canBeHost := user.Id == room.CreatedBy && room.Host == nil

		if full && !canBeHost {
			return custerr.NewConflictErr("the room is already full")
		}

		if canBeHost {
			room.Host = &domain.Host{User: user}
		} else {
			room.Players = append(room.Players, domain.Player{User: user})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(id)
	roomServerChannel := s.roomChannelGetter.Get(domain.ROOM_PREFIX + id)
	if err := roomServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return nil, err
	}
	lobbyServerChannel := s.lobbyChannelGetter.Get(domain.LOBBY)
	if err := lobbyServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return nil, err
	}
	return newRoom.GetProjection(user.Id), nil
}

func (s *RoomService) Connect(ctx context.Context, userId, id string) (*domain.Room, error) {
	if err := s.roomCache.Persist(ctx, id); err != nil {
		return nil, err
	}

	return s.roomCache.SafeSet(ctx, id, func(room *domain.Room) error {
		if room.IsUserHost(userId) {
			room.Host.IsConnected = true
		} else {
			i := slices.IndexFunc(room.Players, func(p domain.Player) bool {
				return p.Id == userId
			})
			room.Players[i].IsConnected = true
		}
		return nil
	})
}

func (s *RoomService) Leave(ctx context.Context, userId, id string) error {
	_, err := s.roomCache.SafeSet(ctx, id, func(room *domain.Room) error {
		if !room.IsUserIn(userId) {
			return custerr.NewForbiddenErr("not allowed to leave room you are not in")
		}
		if room.State != domain.WaitingForStart && room.State != domain.GameOver {
			return custerr.NewConflictErr("cannot leave ongoing game")
		}

		if room.IsUserHost(userId) {
			room.Host = nil
		} else {
			room.Players = slices.DeleteFunc(room.Players, func(p domain.Player) bool {
				return p.Id == userId
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(id)
	roomServerChannel := s.roomChannelGetter.Get(domain.ROOM_PREFIX + id)
	if err := roomServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}
	lobbyServerChannel := s.lobbyChannelGetter.Get(domain.LOBBY)
	return lobbyServerChannel.Send(ctx, roomUpdatedMessage)
}

func (s *RoomService) GetHistory(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.Room, int, error) {
	return s.roomRepository.GetByParticipant(ctx, userId, search)
}
