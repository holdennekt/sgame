package app

import (
	"context"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	_ "github.com/holdennekt/sgame/backend/docs"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/transport/http"
	"github.com/holdennekt/sgame/backend/internal/transport/ws"
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
	authController                    *http.AuthController
	userController                    *http.UserController
	packController                    *http.PackController
	packDraftController               *http.PackDraftController
	roomController                    *http.RoomController
	lobbyHandler                      *ws.LobbyHandler
	roomHandler                       *ws.RoomHandler
	roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter
}

func NewApp(roomCache cache.Room, authController *http.AuthController, userController *http.UserController, packController *http.PackController, packDraftController *http.PackDraftController, roomController *http.RoomController, lobbyHandler *ws.LobbyHandler, roomHandler *ws.RoomHandler, roomInternalEventsProcessorGetter eventsprocessor.RoomInternalEventsProcessorGetter) *app {
	return &app{roomCache, authController, userController, packController, packDraftController, roomController, lobbyHandler, roomHandler, roomInternalEventsProcessorGetter}
}

func (a *app) Run() {
	rooms, _ := a.roomCache.Get(context.Background())
	for _, room := range rooms {
		ok, err := a.roomCache.TrySetOwner(context.Background(), room.Id, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			continue
		}

		processor, err := a.roomInternalEventsProcessorGetter(room.Id)
		go processor.Listen(context.Background())
	}

	go a.roomCache.ListenForExpiredOwners(context.Background(), func(roomId string) {
		if _, err := a.roomCache.GetById(context.Background(), roomId); err != nil {
			return
		}

		ok, err := a.roomCache.TrySetOwner(context.Background(), roomId, eventsprocessor.OWNER_TTL)
		if err != nil || !ok {
			return
		}

		processor, err := a.roomInternalEventsProcessorGetter(roomId)
		if err != nil {
			log.Println("Error while creation roomInternalEventsProcessor:", err)
			return
		}
		go processor.Listen(context.Background())
	})

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
		http.LoggingMiddleware,
		cors.New(corsConfig),
		http.ErrorMiddleware,
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
	log.Fatal(engine.Run(servAddres))
}
