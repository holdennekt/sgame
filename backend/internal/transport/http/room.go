package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/internal/eventsprocessor"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/service"
)

const PASSWORD_QUERY_PARAM = "password"

type RoomController struct {
	serverChannelGetter realtime.ServerChannelGetter
	roomCache           cache.Room
	roomRepository      repository.Room
	packService         *service.PackService
	roomService         *service.RoomService
}

func NewRoomController(serverChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, roomRepository repository.Room, packService *service.PackService, roomService *service.RoomService) *RoomController {
	return &RoomController{serverChannelGetter, roomCache, roomRepository, packService, roomService}
}

func (c *RoomController) RegisterRoutes(r *gin.RouterGroup) {
	rooms := r.Group("/rooms")
	rooms.POST("/", c.create)
	rooms.GET("/", c.get)
	rooms.GET("/:id", c.getProjection)
	rooms.PATCH("/:id/join", c.join)
	rooms.PATCH("/:id/leave", c.leave)
}

func (c *RoomController) create(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)

	var roomDTO domain.RoomDTO
	if err := ctx.ShouldBindJSON(&roomDTO); err != nil {
		ctx.Error(err)
		return
	}

	id, err := c.roomService.Create(ctx, dto.CreateRoomDTO{UserId: userId, RoomDTO: &roomDTO})
	if err != nil {
		ctx.Error(err)
		return
	}
	pack, err := c.packService.GetById(ctx, dto.GetPackByIdDTO{Id: roomDTO.PackId})
	if err != nil {
		ctx.Error(err)
		return
	}

	processor := eventsprocessor.NewRoomEventsProcessor(
		nil,
		c.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel),
		c.serverChannelGetter.Get(domain.ROOM_PREFIX+id).(realtime.Channel),
		c.serverChannelGetter.Get(domain.ROOM_PREFIX+id+domain.INTERNAL_POSTFIX).(realtime.Channel),
		c.roomCache,
		c.roomRepository,
		id,
		domain.User{},
		pack,
	)
	go processor.ListenInternal(context.Background())

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (c *RoomController) get(ctx *gin.Context) {
	rooms, err := c.roomService.Get(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, rooms)
}

func (c *RoomController) getProjection(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")
	password := ctx.Query(PASSWORD_QUERY_PARAM)

	room, err := c.roomService.GetProjection(
		ctx,
		dto.GetRoomProjectionDTO{UserId: userId, Id: id, Password: password},
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, room)
}

func (c *RoomController) join(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")
	password := ctx.Query(PASSWORD_QUERY_PARAM)

	room, err := c.roomService.Join(
		ctx,
		dto.GetRoomProjectionDTO{UserId: userId, Id: id, Password: password},
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, room)
}

func (c *RoomController) leave(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")

	err := c.roomService.Leave(ctx, dto.ConnectRoomDTO{UserId: userId, Id: id})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusOK)
}
