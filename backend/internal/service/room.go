package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/pkg/custerr"
)

const UPDATE_ROOM_RETRIES = 3

type RoomService struct {
	serverChannelGetter realtime.ServerChannelGetter
	roomCache           cache.Room
	userRepository      repository.User
	packRepository      repository.Pack
	roomRepository      repository.Room
}

func NewRoomService(serverChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, userRepository repository.User, packRepository repository.Pack, roomRepository repository.Room) *RoomService {
	return &RoomService{serverChannelGetter, roomCache, userRepository, packRepository, roomRepository}
}

func (s *RoomService) Create(ctx context.Context, dto dto.CreateRoomDTO) (string, error) {
	pack, err := s.packRepository.GetById(ctx, dto.RoomDTO.PackId)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}

	room := &domain.Room{
		Id:      id.String(),
		RoomDTO: *dto.RoomDTO,
		PackPreview: domain.PackPreview{
			Id:   pack.Id,
			Name: pack.Name,
		},
		CreatedBy: dto.UserId,
		Players:   make([]domain.Player, 0),
		State:     domain.WaitingForStart,
	}

	if err := s.roomCache.Set(ctx, room); err != nil {
		return "", custerr.NewInternalErr(err)
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(room.Id)
	roomServerChannel := s.serverChannelGetter.Get(domain.ROOM_PREFIX + room.Id).(realtime.Channel)
	if err := roomServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return "", err
	}
	lobbyServerChannel := s.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel)
	if err := lobbyServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return "", err
	}
	return room.Id, nil
}

func (s *RoomService) GetById(ctx context.Context, id string) (*domain.Room, error) {
	return s.roomCache.GetById(ctx, id)
}

func (s *RoomService) GetProjection(ctx context.Context, dto dto.GetRoomProjectionDTO) (any, error) {
	room, err := s.roomCache.GetById(ctx, dto.Id)
	if err != nil {
		return nil, err
	}

	if room.Options.Type == domain.Private && *room.Options.Password != dto.Password {
		return nil, custerr.NewForbiddenErr("wrong password")
	}

	return room.GetProjection(dto.UserId), nil
}

func (s *RoomService) Get(ctx context.Context) ([]domain.RoomLobby, error) {
	return s.roomCache.Get(ctx)
}

func (s *RoomService) Join(ctx context.Context, dto dto.GetRoomProjectionDTO) (any, error) {
	newRoom, err := s.roomCache.SafeSet(ctx, dto.Id, func(room *domain.Room) error {
		if room.IsUserIn(dto.UserId) {
			return nil
		}
		if room.IsUserBanned(dto.UserId) {
			return custerr.NewForbiddenErr("you were banned from this room")
		}
		if room.Options.Type == domain.Private && *room.Options.Password != dto.Password {
			return custerr.NewForbiddenErr("wrong password")
		}

		full := len(room.Players) >= room.Options.MaxPlayers
		canBeHost := dto.UserId == room.CreatedBy && room.Host == nil

		if full && !canBeHost {
			return custerr.NewConflictErr("the room is already full")
		}

		user, err := s.userRepository.GetById(ctx, dto.UserId)
		if err != nil {
			return err
		}

		if canBeHost {
			room.Host = &domain.Host{User: user.User}
		} else {
			room.Players = append(room.Players, domain.Player{User: user.User})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(dto.Id)
	roomServerChannel := s.serverChannelGetter.Get(domain.ROOM_PREFIX + dto.Id).(realtime.Channel)
	if err := roomServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return nil, err
	}
	lobbyServerChannel := s.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel)
	if err := lobbyServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return nil, err
	}
	return newRoom.GetProjection(dto.UserId), nil
}

func (s *RoomService) Connect(ctx context.Context, dto dto.ConnectRoomDTO) (*domain.Room, error) {
	if err := s.roomCache.Persist(ctx, dto.Id); err != nil {
		return nil, err
	}

	return s.roomCache.SafeSet(ctx, dto.Id, func(room *domain.Room) error {
		if room.IsUserHost(dto.UserId) {
			room.Host.IsConnected = true
		} else {
			i := slices.IndexFunc(room.Players, func(p domain.Player) bool {
				return p.Id == dto.UserId
			})
			room.Players[i].IsConnected = true
		}
		return nil
	})
}

func (s *RoomService) Leave(ctx context.Context, dto dto.ConnectRoomDTO) error {
	_, err := s.roomCache.SafeSet(ctx, dto.Id, func(room *domain.Room) error {
		if !room.IsUserIn(dto.UserId) {
			return custerr.NewForbiddenErr("not allowed to leave room you are not in")
		}
		if room.State != domain.WaitingForStart {
			return custerr.NewConflictErr("cannot leave ongoing game")
		}

		if room.IsUserHost(dto.UserId) {
			room.Host = nil
		} else {
			room.Players = slices.DeleteFunc(room.Players, func(p domain.Player) bool {
				return p.Id == dto.UserId
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(dto.Id)
	roomServerChannel := s.serverChannelGetter.Get(domain.ROOM_PREFIX + dto.Id).(realtime.Channel)
	if err := roomServerChannel.Send(ctx, roomUpdatedMessage); err != nil {
		return err
	}
	lobbyServerChannel := s.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel)
	return lobbyServerChannel.Send(ctx, roomUpdatedMessage)
}
