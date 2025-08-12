package ws

import (
	"context"
	"fmt"
	"log"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/eventsprocessor"
	"github.com/holdennekt/sgame/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/internal/infrastructure/realtime/ws"
	"github.com/holdennekt/sgame/internal/interface/cache"
	"github.com/holdennekt/sgame/internal/interface/realtime"
	"github.com/holdennekt/sgame/internal/service"
	"github.com/holdennekt/sgame/internal/transport/http"
	"github.com/holdennekt/sgame/pkg/custerr"
)

type LobbyHandler struct {
	serverChannelGetter realtime.ServerChannelGetter
	roomCache           cache.Room
	userService         *service.UserService
}

func NewLobbyHandler(serverChannelGetter realtime.ServerChannelGetter, roomCache cache.Room, userService *service.UserService) *LobbyHandler {
	return &LobbyHandler{serverChannelGetter, roomCache, userService}
}

func (h *LobbyHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/lobby", h.connect)
}

func (h *LobbyHandler) connect(ctx *gin.Context) {
	userId := ctx.MustGet(http.USER_ID_CONTEXT_KEY).(string)

	user, err := h.userService.GetById(ctx, userId)
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

	clientChannel := ws.NewChannel(conn)
	serverChannel := h.serverChannelGetter.Get(domain.LOBBY).(realtime.Channel)

	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has connected", user.Name))
	if err := clientChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}
	if err := serverChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}

	processor := eventsprocessor.NewLobbyEventsProcessor(
		clientChannel,
		serverChannel,
		h.roomCache,
		*user,
	)
	go processor.Listen(context.Background())
}
