package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/service"
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

func (c *UserController) getFromSession(ctx *gin.Context) {
	id := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)

	user, err := c.userService.GetById(ctx, id)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) create(ctx *gin.Context) {
	var user domain.DbUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.Error(err)
		return
	}

	id, err := c.userService.Create(ctx, &user)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (c *UserController) getById(ctx *gin.Context) {
	id := ctx.Param("id")

	user, err := c.userService.GetById(ctx, id)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) update(ctx *gin.Context) {
	id := ctx.Param("id")

	var user domain.DbUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.Error(err)
		return
	}
	user.Id = id

	if err := c.userService.Update(ctx, &user); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *UserController) delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.userService.Delete(ctx, id); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
