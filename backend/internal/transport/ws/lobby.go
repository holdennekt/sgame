package ws

import (
	"context"
	"fmt"
	"log"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	_ "github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/ws"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/transport/http"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type LobbyHandler struct {
	lobbyChannelGetter         realtime.ServerChannelGetter
	lobbyEventsProcessorGetter eventsprocessor.LobbyEventsProcessorGetter
	shutdownCtx                context.Context
}

func NewLobbyHandler(lobbyChannelGetter realtime.ServerChannelGetter, lobbyEventsProcessorGetter eventsprocessor.LobbyEventsProcessorGetter) *LobbyHandler {
	return &LobbyHandler{lobbyChannelGetter: lobbyChannelGetter, lobbyEventsProcessorGetter: lobbyEventsProcessorGetter, shutdownCtx: context.Background()}
}

func (h *LobbyHandler) SetShutdownCtx(ctx context.Context) {
	h.shutdownCtx = ctx
}

// @Summary      Connect to Lobby WebSocket
// @Description  Establishes a WebSocket connection to the global lobby.
// @Description  Requires a valid session cookie. Once connected, sends a chat message "{User} has connected".
// @Tags         lobby
// @Success      101  {string}  string "Switching Protocols"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /lobby [get]
func (h *LobbyHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/lobby", h.connect)
}

func (h *LobbyHandler) connect(ctx *gin.Context) {
	user := ctx.MustGet(http.USER_CONTEXT_KEY).(domain.User)

	conn, err := websocket.Accept(ctx.Writer, ctx.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		// CompressionMode:    websocket.CompressionContextTakeover,
	})
	if err != nil {
		ctx.Error(custerr.NewInternalErr(err))
		return
	}

	clientChannel := ws.NewChannel(conn)
	serverChannel := h.lobbyChannelGetter.Get(domain.LOBBY)

	chatMessage := client.NewSystemChatMessage(fmt.Sprintf("%s has connected", user.Name))
	if err := clientChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}
	if err := serverChannel.Send(ctx, chatMessage); err != nil {
		log.Println(err)
		return
	}

	processor := h.lobbyEventsProcessorGetter(
		clientChannel,
		user,
	)
	go processor.Listen(h.shutdownCtx)
}
