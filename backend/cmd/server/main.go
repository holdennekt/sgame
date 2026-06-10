package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/holdennekt/sgame/backend/internal/app"
	"github.com/holdennekt/sgame/backend/internal/config"
	gcsStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/gcs"
	minioStorage "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/minio"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custvalid"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
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

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	if cfg.AppEnv == "production" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	}

	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s?authSource=admin",
		cfg.MongoUser,
		cfg.MongoPassword,
		cfg.MongoHost,
		cfg.MongoPort,
		cfg.MongoName,
	)
	opts := options.Client().ApplyURI(mongoURI)
	conn, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		slog.Error("failed to connect to mongodb", "err", err)
		os.Exit(1)
	}
	defer conn.Disconnect(context.Background())

	mdb := conn.Database(cfg.MongoName)

	rds := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Username: cfg.RedisUser,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rds.Close()

	var store storage.Storage

	if cfg.StorageProvider == "gcs" {
		store, err = gcsStorage.NewGCSStorage(context.Background(), cfg.BucketName)
		if err != nil {
			slog.Error("failed to initialize gcs storage", "err", err)
			os.Exit(1)
		}
	} else {
		mio, err := minio.New(cfg.MinioEndpoint, &minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.MinioUser,
				cfg.MinioPassword,
				"",
			),
			Secure: cfg.MinioUseSSL,
		})
		if err != nil {
			slog.Error("failed to initialize minio client", "err", err)
			os.Exit(1)
		}
		store = minioStorage.NewMinioStorage(mio, cfg.BucketName, cfg.FrontendURL, cfg.UserAgent)
	}

	app := app.InitializeApp(mdb, rds, store, cfg)
	app.Run()
}
