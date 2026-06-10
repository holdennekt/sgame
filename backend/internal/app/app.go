package app

import (
	"context"
	"errors"
	"log"
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
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	myHttp "github.com/holdennekt/sgame/backend/internal/transport/http"
	myWs "github.com/holdennekt/sgame/backend/internal/transport/ws"
	"github.com/holdennekt/sgame/backend/pkg/custvalid"
	"github.com/holdennekt/sgame/backend/pkg/envvar"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	log.SetFlags(0)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation(custvalid.SameLength, custvalid.ValidateSameLength)
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

func NewApp(roomCache cache.Room, authController *myHttp.AuthController, userController *myHttp.UserController, packController *myHttp.PackController, packDraftController *myHttp.PackDraftController, roomController *myHttp.RoomController, lobbyHandler *myWs.LobbyHandler, roomHandler *myWs.RoomHandler, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter) *app {
	return &app{roomCache, authController, userController, packController, packDraftController, roomController, lobbyHandler, roomHandler, roomInternalEventsProcessorGetter}
}

func (a *app) Run() {
	appCtx, appCancel := context.WithCancel(context.Background())

	rooms, _ := a.roomCache.Get(appCtx)
	for _, room := range rooms {
		ok, err := a.roomCache.TrySetOwner(appCtx, room.Id, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			continue
		}

		processor, err := a.roomInternalEventsProcessorGetter(room.Id)
		go processor.Listen(appCtx)
	}

	go a.roomCache.ListenForExpiredOwners(appCtx, func(roomId string) {
		if _, err := a.roomCache.GetById(appCtx, roomId); err != nil {
			return
		}

		ok, err := a.roomCache.TrySetOwner(appCtx, roomId, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			return
		}

		processor, err := a.roomInternalEventsProcessorGetter(roomId)
		if err != nil {
			log.Println("Error while creation roomInternalEventsProcessor:", err)
			return
		}
		go processor.Listen(appCtx)
	})

	a.lobbyHandler.SetShutdownCtx(appCtx)
	a.roomHandler.SetShutdownCtx(appCtx)

	corsConfig := cors.DefaultConfig()
	if os.Getenv("APP_ENV") == "development" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{envvar.GetEnvVar("FRONTEND_URL")}
	}
	corsConfig.AllowCredentials = true

	engine := gin.New()

	engine.GET("api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.Use(
		gin.Recovery(),
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

	servAddres := envvar.GetEnvVar("HOST") + ":" + envvar.GetEnvVar("PORT")

	srv := &http.Server{
		Addr:    servAddres,
		Handler: engine,
	}

	go func() {
		log.Println("Starting server at addres:", servAddres)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	appCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}

	log.Println("Server has been shutdown gracefully")
}
