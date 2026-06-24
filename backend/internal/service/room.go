package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/backend/internal/config"
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
	cfg                               *config.Config
}

func NewRoomService(packRepository repository.Pack, roomRepository repository.Room, roomCache cache.Room, lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ServerChannelGetter, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter, cfg *config.Config) *RoomService {
	return &RoomService{packRepository, roomRepository, roomCache, lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter, roomInternalEventsProcessorGetter, cfg}
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

	options := crr.Options
	options.TimeToBet = s.cfg.TimeToBet
	options.TimeToPass = s.cfg.TimeToPass

	room := &domain.Room{
		Id:   id.String(),
		Name: crr.Name,
		PackPreview: domain.PackPreview{
			Id:   pack.Id,
			Name: pack.Name,
		},
		Options:   options,
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

	if !room.IsUserIn(userId) {
		if room.IsUserBanned(userId) {
			return nil, custerr.NewForbiddenErr("you were banned from this room")
		}
		if room.Options.Type == domain.Private && *room.Options.Password != password {
			return nil, custerr.NewForbiddenErr("wrong password")
		}
	}

	spectatorCount, _ := s.roomCache.GetSpectatorCount(ctx, id)
	return room.GetProjection(userId, spectatorCount), nil
}

func (s *RoomService) Get(ctx context.Context) ([]domain.RoomLobby, error) {
	return s.roomCache.Get(ctx)
}

func (s *RoomService) GetSpectatorCount(ctx context.Context, id string) (int, error) {
	return s.roomCache.GetSpectatorCount(ctx, id)
}

func (s *RoomService) GetHistory(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.Room, int, error) {
	return s.roomRepository.GetByParticipant(ctx, userId, search)
}

func (s *RoomService) Join(ctx context.Context, user domain.User, id, password string) (any, error) {
	room, err := s.roomCache.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if room.IsUserIn(user.Id) {
		spectatorCount, _ := s.roomCache.GetSpectatorCount(ctx, id)
		return room.GetProjection(user.Id, spectatorCount), nil
	}

	newRoom, err := s.roomCache.SafeUpdate(ctx, id, func(room *domain.Room) error {
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
	spectatorCount, _ := s.roomCache.GetSpectatorCount(ctx, id)
	return newRoom.GetProjection(user.Id, spectatorCount), nil
}

func (s *RoomService) Leave(ctx context.Context, userId, id string) error {
	_, err := s.roomCache.SafeUpdate(ctx, id, func(room *domain.Room) error {
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

func (s *RoomService) Connect(ctx context.Context, userId, roomId string) (*domain.Room, error) {
	room, err := s.roomCache.GetById(ctx, roomId)
	if err != nil {
		return nil, err
	}

	if !room.IsUserIn(userId) {
		_, err := s.roomCache.IncrSpectators(ctx, roomId)
		return room, err
	}

	if err := s.roomCache.Persist(ctx, roomId); err != nil {
		return nil, err
	}

	return s.roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.IsUserHost(userId) {
			room.Host.IsConnected = true
		} else {
			playerIndex := room.UsersPlayerIndex(userId)
			room.Players[playerIndex].IsConnected = true
		}
		return nil
	})
}

func (s *RoomService) Disconnect(ctx context.Context, userId, roomId string) (*domain.Room, error) {
	room, err := s.roomCache.GetById(ctx, roomId)
	if err != nil {
		if _, ok := err.(custerr.NotFoundErr); ok {
			return nil, nil
		}
		return nil, err
	}

	if !room.IsUserIn(userId) {
		_, err := s.roomCache.DecrSpectators(ctx, roomId)
		return room, err
	}

	return s.roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
		if room.IsUserHost(userId) {
			room.Host.IsConnected = false
		} else {
			playerIndex := room.UsersPlayerIndex(userId)
			room.Players[playerIndex].IsConnected = false
		}
		return nil
	})
}
