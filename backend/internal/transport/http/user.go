package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService}
}

func (c *UserController) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/user", c.getFromSession)
	users := r.Group("/users")
	users.POST("/", c.create)
	users.GET("/:id", c.getById)
	users.PUT("/:id", c.update)
	users.DELETE("/:id", c.delete)
}

// @Summary      Get current user
// @Description  Retrieves the profile of the currently authenticated user based on session
// @Tags         users
// @Produce      json
// @Success      200  {object}  domain.DbUser
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or invalid"
// @Failure      404  {object}  dto.ErrorResponse "User not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /user [get]
func (c *UserController) getFromSession(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)
	ctx.JSON(http.StatusOK, user)
}

// @Summary      Create a new user
// @Description  Creates a new user account with provided data
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user body      domain.DbUser true "User data"
// @Success      201  {object}  dto.CreateUserResponse
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /users [post]
func (c *UserController) create(ctx *gin.Context) {
	var user domain.DbUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		_ = ctx.Error(err)
		return
	}

	id, err := c.userService.Create(ctx, &user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.CreateUserResponse{Id: id})
}

// @Summary      Get user by ID
// @Description  Retrieves a user profile by their unique identifier
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  domain.DbUser
// @Failure      404  {object}  dto.ErrorResponse "User not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /users/{id} [get]
func (c *UserController) getById(ctx *gin.Context) {
	id := ctx.Param("id")

	user, err := c.userService.GetById(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// @Summary      Update user
// @Description  Updates an existing user's information by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string         true "User ID"
// @Param        user  body      domain.DbUser  true "Updated user data"
// @Success      204   "No Content"
// @Failure      400   {object}  dto.ErrorResponse "Invalid input data"
// @Failure      404   {object}  dto.ErrorResponse "User not found"
// @Failure      500   {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /users/{id} [put]
func (c *UserController) update(ctx *gin.Context) {
	requesterId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")
	if requesterId != id {
		_ = ctx.Error(custerr.NewForbiddenErr("not allowed to update other users"))
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	user := &domain.DbUser{
		User:     domain.User{Id: id, Name: req.Name, Avatar: req.Avatar},
		Password: req.Password,
	}
	if err := c.userService.Update(ctx, user); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// @Summary      Delete user
// @Description  Permanently removes a user account by ID
// @Tags         users
// @Param        id   path      string  true  "User ID"
// @Success      204  "No Content"
// @Failure      404  {object}  dto.ErrorResponse "User not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /users/{id} [delete]
func (c *UserController) delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.userService.Delete(ctx, id); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
