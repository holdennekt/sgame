package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	_ "github.com/holdennekt/sgame/backend/docs"
	"github.com/holdennekt/sgame/backend/internal/config"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	myHttp "github.com/holdennekt/sgame/backend/internal/transport/http"
	myWs "github.com/holdennekt/sgame/backend/internal/transport/ws"
	"github.com/holdennekt/sgame/backend/pkg/custvalid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation(custvalid.SameLength, custvalid.ValidateSameLength)
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				return fld.Name
			}
			return name
		})
	}
}

type app struct {
	cfg                               *config.Config
	roomCache                         cache.Room
	authController                    *myHttp.AuthController
	userController                    *myHttp.UserController
	packController                    *myHttp.PackController
	packDraftController               *myHttp.PackDraftController
	roomController                    *myHttp.RoomController
	lobbyHandler                      *myWs.LobbyHandler
	roomHandler                       *myWs.RoomHandler
	roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter
}

func NewApp(cfg *config.Config, roomCache cache.Room, authController *myHttp.AuthController, userController *myHttp.UserController, packController *myHttp.PackController, packDraftController *myHttp.PackDraftController, roomController *myHttp.RoomController, lobbyHandler *myWs.LobbyHandler, roomHandler *myWs.RoomHandler, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter) *app {
	return &app{cfg, roomCache, authController, userController, packController, packDraftController, roomController, lobbyHandler, roomHandler, roomInternalEventsProcessorGetter}
}

// Start sets up background goroutines and returns the HTTP handler.
// The caller is responsible for canceling ctx to stop background work.
func (a *app) Start(ctx context.Context) http.Handler {
	rooms, _ := a.roomCache.Get(ctx)
	for _, room := range rooms {
		ok, err := a.roomCache.TrySetOwner(ctx, room.Id, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			continue
		}

		processor, err := a.roomInternalEventsProcessorGetter(room.Id)
		if err != nil {
			slog.Error("error creating room internal events processor", "err", err)
			continue
		}
		go processor.Listen(ctx)
	}

	go a.roomCache.ListenForExpiredOwners(ctx, func(roomId string) {
		if _, err := a.roomCache.GetById(ctx, roomId); err != nil {
			return
		}

		ok, err := a.roomCache.TrySetOwner(ctx, roomId, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			return
		}

		processor, err := a.roomInternalEventsProcessorGetter(roomId)
		if err != nil {
			slog.Error("error creating room internal events processor", "err", err)
			return
		}
		go processor.Listen(ctx)
	})

	a.lobbyHandler.SetShutdownCtx(ctx)
	a.roomHandler.SetShutdownCtx(ctx)

	corsConfig := cors.DefaultConfig()
	if a.cfg.AppEnv == "development" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{a.cfg.FrontendURL}
	}
	corsConfig.AllowCredentials = true

	engine := gin.New()

	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.GET("api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.Use(
		gin.Recovery(),
		myHttp.RequestIDMiddleware,
		myHttp.PrometheusMiddleware,
		myHttp.LoggingMiddleware,
		cors.New(corsConfig),
		myHttp.ErrorMiddleware,
	)

	api := engine.Group("/api")
	a.authController.RegisterRoutes(api)

	protected := api.Group("/", a.authController.Authorize)
	a.userController.RegisterRoutes(protected)
	a.packController.RegisterRoutes(protected)
	a.packDraftController.RegisterRoutes(protected)
	a.roomController.RegisterRoutes(protected)

	wsGroup := protected.Group("/ws")
	a.lobbyHandler.RegisterRoute(wsGroup)
	a.roomHandler.RegisterRoute(wsGroup)

	return engine
}

func (a *app) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	handler := a.Start(ctx)

	servAddres := a.cfg.Host + ":" + a.cfg.Port
	srv := &http.Server{
		Addr:    servAddres,
		Handler: handler,
	}

	go func() {
		slog.Info("starting server", "addr", servAddres)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown gracefully", "err", err)
		os.Exit(1)
	}

	slog.Info("server shutdown gracefully")
}
