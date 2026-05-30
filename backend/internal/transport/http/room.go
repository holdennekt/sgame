package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
)

const PASSWORD_QUERY_PARAM = "password"

type RoomController struct {
	packService *service.PackService
	roomService *service.RoomService
}

func NewRoomController(packService *service.PackService, roomService *service.RoomService) *RoomController {
	return &RoomController{packService, roomService}
}

func (c *RoomController) RegisterRoutes(r *gin.RouterGroup) {
	rooms := r.Group("/rooms")
	rooms.POST("/", c.create)
	rooms.GET("/", c.get)
	rooms.GET("/history", c.getHistory)
	rooms.GET("/:id", c.getProjection)
	rooms.PATCH("/:id/join", c.join)
	rooms.PATCH("/:id/leave", c.leave)
}

// @Summary      Create a new room
// @Description  Creates a game room for the authenticated user
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateRoomRequest true "Room creation data"
// @Success      201  {object}  dto.CreateRoomResponse
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms [post]
func (c *RoomController) create(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id

	var crr dto.CreateRoomRequest
	if err := ctx.ShouldBindJSON(&crr); err != nil {
		ctx.Error(err)
		return
	}

	id, err := c.roomService.Create(ctx, userId, crr)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.CreateRoomResponse{Id: id})
}

// @Summary      List all rooms
// @Description  Returns a list of all available game rooms
// @Tags         rooms
// @Produce      json
// @Success      200  {array}   domain.RoomLobby
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms [get]
func (c *RoomController) get(ctx *gin.Context) {
	rooms, err := c.roomService.Get(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, rooms)
}

// @Summary      Get room projection
// @Description  Retrieves a specific room's state (projection) by ID. Used for initial load.
// @Tags         rooms
// @Produce      json
// @Param        id        path      string  true   "Room ID"
// @Param        password  query     string  false  "Room password (if required)"
// @Success      200  {object}  domain.RoomLobby "Can be RoomLobby, RoomHost, or RoomPlayer"
// @Success      200  {object}  domain.RoomHost "Can be RoomLobby, RoomHost, or PlayerRoom"
// @Success      200  {object}  domain.RoomPlayer "Can be RoomLobby, HostRoom, or RoomPlayer"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Invalid password"
// @Failure      404  {object}  dto.ErrorResponse "Room not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms/{id} [get]
func (c *RoomController) getProjection(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")
	password := ctx.Query(PASSWORD_QUERY_PARAM)

	room, err := c.roomService.GetProjection(ctx, userId, id, password)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, room)
}

// @Summary      Join room
// @Description  Adds the authenticated user to a room
// @Tags         rooms
// @Produce      json
// @Param        id        path      string  true   "Room ID"
// @Param        password  query     string  false  "Room password (if required)"
// @Success      200  {object}  domain.Room
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Room full or invalid password"
// @Failure      404  {object}  dto.ErrorResponse "Room not found"
// @Failure      409  {object}  dto.ErrorResponse "Room already full"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms/{id}/join [patch]
func (c *RoomController) join(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)
	id := ctx.Param("id")
	password := ctx.Query(PASSWORD_QUERY_PARAM)

	room, err := c.roomService.Join(ctx, user, id, password)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, room)
}

// @Summary      Leave room
// @Description  Removes the authenticated user from a room
// @Tags         rooms
// @Param        id   path      string  true  "Room ID"
// @Success      200  "OK"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404  {object}  dto.ErrorResponse "Room not found"
// @Failure      409  {object}  dto.ErrorResponse "Cannot leave ongoing game"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms/{id}/leave [patch]
func (c *RoomController) leave(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	err := c.roomService.Leave(ctx, userId, id)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusOK)
}

// @Summary      Get room history
// @Description  Returns a paginated list of past rooms the authenticated user has participated in
// @Tags         rooms
// @Produce      json
// @Param        query query     dto.SearchRequest false "Pagination and search parameters"
// @Success      200  {object}  dto.SearchResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /rooms/history [get]
func (c *RoomController) getHistory(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id

	var query dto.SearchRequest
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.Error(err)
		return
	}
	if query.Page == 0 {
		query.Page = DEFAULT_PAGE
	}
	if query.Limit == 0 {
		query.Limit = DEFAULT_LIMIT
	}

	rooms, total, err := c.roomService.GetHistory(ctx, userId, query)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SearchResponse{
		Items:    rooms,
		Total:    total,
		Page:     query.Page,
		PageSize: query.Limit,
		HasNext:  query.Page*query.Limit < total,
	})
}
