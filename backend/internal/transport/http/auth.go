package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

const SESSION_ID_COOKIE_NAME = "sessionId"
const USER_CONTEXT_KEY = "user"
const SESSION_COOKIE_TTL = 7 * 24 * 60 * 60

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService}
}

func (c *AuthController) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", c.login)
	r.POST("/register", c.register)
	r.DELETE("/logout", c.logout)
	r.POST("/guest", c.guest)
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
		_ = ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Login(ctx, cur)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, SESSION_COOKIE_TTL, "", "", false, true)
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
		_ = ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.Register(ctx, cur)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, SESSION_COOKIE_TTL, "", "", false, true)
	ctx.JSON(http.StatusOK, dto.AuthResponse{UserId: userId})
}

// @Summary      User Logout
// @Description  Invalidates the current session and clears the session cookie
// @Tags         auth
// @Success      204  "No Content"
// @Failure      500  {object}  dto.ErrorResponse "Internal Server Error"
// @Security     CookieAuth
// @Router       /logout [delete]
func (c *AuthController) logout(ctx *gin.Context) {
	sessionId, err := ctx.Cookie(SESSION_ID_COOKIE_NAME)
	if err != nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	_ = c.authService.Logout(ctx, sessionId)
	ctx.SetCookie(SESSION_ID_COOKIE_NAME, "", -1, "", "", false, true)
	ctx.Status(http.StatusNoContent)
}

// @Summary      Guest Login
// @Description  Creates a temporary guest session using a display name, sets an HttpOnly cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.GuestLoginRequest true "Guest login data"
// @Success      200  {object}  dto.AuthResponse "Successfully logged in as guest"
// @Header       200  {string}  Set-Cookie "session_id=abc...; HttpOnly; Path=/"
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      500  {object}  dto.ErrorResponse "Internal Server Error"
// @Router       /guest [post]
func (c *AuthController) guest(ctx *gin.Context) {
	var req dto.GuestLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	sessionId, userId, err := c.authService.GuestLogin(ctx, req.Name)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, SESSION_COOKIE_TTL, "", "", false, true)
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

	user, err := c.authService.GetUser(ctx, sessionId)
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

	ctx.Set(USER_CONTEXT_KEY, *user)
}
