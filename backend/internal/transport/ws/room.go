package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/internal/eventsprocessor"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/internal/infrastructure/realtime/ws"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/interface/repository"
	"github.com/holdennekt/sgame/internal/message"
	"github.com/holdennekt/sgame/internal/service"
	"github.com/holdennekt/sgame/internal/transport/http"
	"github.com/holdennekt/sgame/pkg/custerr"
)

type RoomHandler struct {
	serverChannelGetter realtime.ServerChannelGetter
	roomCache           cache.Room
	roomRepository      repository.Room
	userService         *service.UserService
	packService         *service.PackService
	roomService         *service.RoomService
}

func NewRoomHandler(serverChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, roomRepository repository.Room, userService *service.UserService, packService *service.PackService, roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{serverChannelGetter, roomCache, roomRepository, userService, packService, roomService}
}

func (h *RoomHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/room/:id", h.connect)
}

func (h *RoomHandler) connect(ctx *gin.Context) {
	userId := ctx.MustGet(http.USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")

	room, err := h.roomService.GetById(ctx, id)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !room.IsUserIn(userId) {
		ctx.Error(custerr.NewForbiddenErr("cannot connect to room you are not in"))
		return
	}

	user, err := h.userService.GetById(ctx, userId)
	if err != nil {
		ctx.Error(err)
		return
	}
	pack, err := h.packService.GetById(ctx, dto.GetPackByIdDTO{Id: room.PackId})
	if err != nil {
		ctx.Error(err)
		return
	}

	conn, err := websocket.Accept(ctx.Writer, ctx.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		// CompressionMode:    websocket.CompressionContextTakeover,
	})
	if err != nil {
		ctx.Error(custerr.NewInternalErr(err))
		return
	}

	newRoom, err := h.roomService.Connect(ctx, dto.ConnectRoomDTO{UserId: userId, Id: id})
	if err != nil {
		log.Println(err)
		return
	}

	clientChannel := ws.NewChannel(conn)
	payload, _ := json.Marshal(newRoom.GetProjection(user.Id))
	clientRoomUpdatedMessage := message.Message{Event: domain.RoomUpdated, Payload: payload}
	if err := clientChannel.Send(ctx, clientRoomUpdatedMessage); err != nil {
		log.Println(err)
		return
	}

	serverRoomUpdatedMessage := outgoing.NewRoomUpdatedMessage(id)
	roomServerChannel := h.serverChannelGetter.Get(domain.ROOM_PREFIX + id).(realtime.Channel)
	if err := roomServerChannel.Send(ctx, serverRoomUpdatedMessage); err != nil {
		log.Println(err)
		return
	}

	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has connected", user.Name))
	if err := clientChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}
	if err := roomServerChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}

	processor := eventsprocessor.NewRoomEventsProcessor(
		clientChannel,
		h.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel),
		roomServerChannel,
		h.serverChannelGetter.Get(domain.ROOM_PREFIX+id+domain.INTERNAL_POSTFIX).(realtime.Channel),
		h.roomCache,
		h.roomRepository,
		id,
		*user,
		pack,
	)
	go processor.Listen(context.Background())
}
