package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
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

// @Summary      User Login
// @Description  Authenticates user, creates a session, and sets an HttpOnly cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserRequest true "Login credentials"
// @Success      200  {object}  dto.AuthResponse "Successfully authenticated"
// @Header       200  {string}  Set-Cookie "session_id=abc...; HttpOnly; Path=/"
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Invalid credentials"
// @Failure      500  {object}  dto.ErrorResponse "Internal Server Error"
// @Router       /login [post]
func (c *AuthController) login(ctx *gin.Context) {
	var cur dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&cur); err != nil {
		ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Login(ctx, cur)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
	ctx.JSON(http.StatusOK, dto.AuthResponse{UserId: userId})
}

// @Summary      User Registration
// @Description  Creates a new account, automatically logs in, and sets an HttpOnly cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserRequest true "Registration data"
// @Success      200  {object}  dto.AuthResponse "Account created and logged in"
// @Header       200  {string}  Set-Cookie "session_id=abc...; HttpOnly; Path=/"
// @Failure      400  {object}  dto.ErrorResponse "User already exists or validation failed"
// @Failure      500  {object}  dto.ErrorResponse "Internal Server Error"
// @Router       /register [post]
func (c *AuthController) register(ctx *gin.Context) {
	var cur dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&cur); err != nil {
		ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Register(ctx, cur)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
	ctx.JSON(http.StatusOK, dto.AuthResponse{UserId: userId})
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
