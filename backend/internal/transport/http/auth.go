package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/service"
	"github.com/holdennekt/sgame/pkg/custerr"
)

const SESSION_ID_COOKIE_NAME = "sessionId"
const USER_ID_CONTEXT_KEY = "userId"

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService}
}

func (c *AuthController) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", c.login)
	r.POST("/register", c.register)
}

func (c *AuthController) login(ctx *gin.Context) {
	var dbUserDTO domain.DbUserDTO
	if err := ctx.ShouldBindJSON(&dbUserDTO); err != nil {
		ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Login(ctx, dbUserDTO)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"userId": userId})
}

func (c *AuthController) register(ctx *gin.Context) {
	var dbUserDTO domain.DbUserDTO
	if err := ctx.ShouldBindJSON(&dbUserDTO); err != nil {
		ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Register(ctx, dbUserDTO)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"userId": userId})
}

func (c *AuthController) Authorize(ctx *gin.Context) {
	sessionId, err := ctx.Cookie(SESSION_ID_COOKIE_NAME)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "missing sessionId cookie"},
		)
		return
	}

	userId, err := c.authService.GetUserID(ctx, sessionId)
	if err != nil {
		switch err := err.(type) {
		case custerr.NotFoundErr:
			ctx.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{"error": err.Error()},
			)
			return
		default:
			ctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)
			return
		}
	}

	ctx.Set(USER_ID_CONTEXT_KEY, userId)
}
