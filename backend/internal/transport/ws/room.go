package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	_ "github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/ws"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/internal/transport/http"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type RoomHandler struct {
	roomService               *service.RoomService
	lobbyChannelGetter        realtime.ServerChannelGetter
	roomChannelGetter         realtime.ServerChannelGetter
	roomInternalChannelGetter realtime.ServerChannelGetter
	roomEventsProcessorGetter eventsprocessor.RoomEventsProcessorGetter
	shutdownCtx               context.Context
}

func NewRoomHandler(roomService *service.RoomService, lobbyChannelGetter, roomChannelGetter, roomInternalChannelGetter realtime.ServerChannelGetter, roomEventsProcessorGetter eventsprocessor.RoomEventsProcessorGetter) *RoomHandler {
	return &RoomHandler{roomService: roomService, lobbyChannelGetter: lobbyChannelGetter, roomChannelGetter: roomChannelGetter, roomInternalChannelGetter: roomInternalChannelGetter, roomEventsProcessorGetter: roomEventsProcessorGetter, shutdownCtx: context.Background()}
}

func (h *RoomHandler) SetShutdownCtx(ctx context.Context) {
	h.shutdownCtx = ctx
}

func (h *RoomHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/room/:id", h.connect)
}

// @Summary      Connect to Room WebSocket
// @Description  Establishes a WebSocket connection to the playing room.
// @Description  Requires a valid session cookie. Once connected, sends a chat message "{User} has connected".
// @Tags         room
// @Success      101  {string}  string "Switching Protocols"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /room/{id} [get]
func (h *RoomHandler) connect(ctx *gin.Context) {
	user := ctx.MustGet(http.USER_CONTEXT_KEY).(domain.User)
	id := ctx.Param("id")

	room, err := h.roomService.GetById(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	if !room.IsUserIn(user.Id) {
		_ = ctx.Error(custerr.NewForbiddenErr("cannot connect to room you are not in"))
		return
	}

	conn, err := websocket.Accept(ctx.Writer, ctx.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		// CompressionMode:    websocket.CompressionContextTakeover,
	})
	if err != nil {
		_ = ctx.Error(custerr.NewInternalErr(err))
		return
	}

	newRoom, err := h.roomService.Connect(ctx, user.Id, id)
	if err != nil {
		slog.Error("error", "err", err)
		return
	}

	clientChannel := ws.NewChannel(conn)
	payload, _ := json.Marshal(newRoom.GetProjection(user.Id))
	clientRoomUpdatedMessage := message.Message{Event: domain.RoomUpdated, Payload: payload}
	if err := clientChannel.Send(ctx, clientRoomUpdatedMessage); err != nil {
		slog.Error("error", "err", err)
		return
	}

	serverRoomUpdatedMessage := outgoing.NewRoomUpdatedMessage(id)
	roomServerChannel := h.roomChannelGetter.Get(domain.ROOM_PREFIX + id)
	if err := roomServerChannel.Send(ctx, serverRoomUpdatedMessage); err != nil {
		slog.Error("error", "err", err)
		return
	}

	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has connected", user.Name))
	if err := clientChannel.Send(ctx, chatMessage); err != nil {
		slog.Error("error", "err", err)
		return
	}
	if err := roomServerChannel.Send(ctx, chatMessage); err != nil {
		slog.Error("error", "err", err)
		return
	}

	processor, err := h.roomEventsProcessorGetter(
		clientChannel,
		id,
		user,
	)
	if err != nil {
		slog.Error("error while creation roomEventsProcessor", "err", err)
		return
	}
	go processor.Listen(h.shutdownCtx)
}
