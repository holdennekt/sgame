package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/holdennekt/sgame/backend/internal/app"
	gcsStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/gcs"
	minioStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/minio"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custvalid"
	"github.com/holdennekt/sgame/backend/pkg/envvar"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	log.SetFlags(0)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation(custvalid.SameLength, custvalid.ValidateSameLength)
	}
}

// @title           SGame API
// @version         1.0
// @description     Backend API for SGame (Go + Next.js).
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name sessionId

func main() {
	godotenv.Load()

	opts := options.Client().ApplyURI(envvar.GetEnvVar("MONGO_CONN_STR"))
	conn, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	mdb := conn.Database(envvar.GetEnvVar("MONGO_NAME"))

	rds := redis.NewClient(&redis.Options{
		Addr:     envvar.GetEnvVar("REDIS_HOST") + ":" + envvar.GetEnvVar("REDIS_PORT"),
		Username: envvar.GetEnvVar("REDIS_USER"),
		Password: envvar.GetEnvVar("REDIS_PASSWORD"),
		DB:       envvar.GetEnvVarInt("REDIS_DB"),
	})
	defer rds.Close()

	var storage storage.Storage

	if envvar.GetEnvVar("STORAGE_PROVIDER") == "gcs" {
		storage, err = gcsStorage.NewGCSStorage(context.Background(), envvar.GetEnvVar("BUCKET_NAME"))
	} else {
		mio, err := minio.New(envvar.GetEnvVar("MINIO_ENDPOINT"), &minio.Options{
			Creds: credentials.NewStaticV4(
				envvar.GetEnvVar("MINIO_ACCESS_KEY"),
				envvar.GetEnvVar("MINIO_SECRET_KEY"),
				"",
			),
			Secure: envvar.GetEnvVarBool("MINIO_USE_SSL"),
		})
		if err != nil {
			log.Fatal(err)
		}
		storage = minioStorage.NewMinioStorage(mio, envvar.GetEnvVar("BUCKET_NAME"), envvar.GetEnvVar("FRONTEND_URL"))
	}

	app := app.InitializeApp(mdb, rds, storage)
	app.Run()
}
