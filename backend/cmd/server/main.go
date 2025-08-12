package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/holdennekt/sgame/internal/transport/http"
	"github.com/holdennekt/sgame/pkg/custvalid"
	"github.com/holdennekt/sgame/pkg/envvar"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation(custvalid.SameLength, custvalid.ValidateSameLength)
	}
}

func main() {
	godotenv.Load()

	rds := redis.NewClient(&redis.Options{
		Addr:     envvar.GetEnvVar("REDIS_HOST") + ":" + envvar.GetEnvVar("REDIS_PORT"),
		Username: envvar.GetEnvVar("REDIS_USERNAME"),
		Password: envvar.GetEnvVar("REDIS_PASSWORD"),
		DB:       envvar.GetEnvVarInt("REDIS_DB"),
	})
	defer rds.Close()

	opts := options.Client().ApplyURI(envvar.GetEnvVar("MONGO_CONN"))
	conn, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	mdb := conn.Database(envvar.GetEnvVar("MONGO_DB_NAME"))

	engine := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{envvar.GetEnvVar("MY_CLIENT_ORIGIN")}
	corsConfig.AllowCredentials = true
	engine.Use(cors.New(corsConfig))

	engine.Use(http.ErrorMiddleware)

	root := engine.Group("/")

	authController := InitializeAuthController(mdb, rds)
	authController.RegisterRoutes(root)

	restGroup := root.Group("/api", authController.Authorize)

	userController := InitializeUserController(mdb)
	userController.RegisterRoutes(restGroup)

	packController := InitializePackController(mdb)
	packController.RegisterRoutes(restGroup)

	roomController := InitializeRoomController(mdb, rds)
	roomController.RegisterRoutes(restGroup)

	wsGroup := root.Group("/ws", authController.Authorize)

	lobbyHandler := InitializeLobbyHandler(mdb, rds)
	lobbyHandler.RegisterRoute(wsGroup)

	roomHandler := InitializeRoomHandler(mdb, rds)
	roomHandler.RegisterRoute(wsGroup)

	servAddres := envvar.GetEnvVar("HOST") + ":" + envvar.GetEnvVar("PORT")
	log.Fatal(engine.Run(servAddres))
}
