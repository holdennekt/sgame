// The attachment-processor runs the raw->compressed pipeline as its own
// service, keeping CPU-heavy transcoding away from live game rooms.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/holdennekt/sgame/backend/internal/config"
	redisCache "github.com/holdennekt/sgame/backend/internal/infrastructure/cache/redis"
	mongoDatabase "github.com/holdennekt/sgame/backend/internal/infrastructure/database/mongo"
	queueFactory "github.com/holdennekt/sgame/backend/internal/infrastructure/queue/factory"
	"github.com/holdennekt/sgame/backend/internal/infrastructure/realtime/pubsub"
	storageFactory "github.com/holdennekt/sgame/backend/internal/infrastructure/storage/factory"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxConcurrentProcessors = 5

func main() {
	_ = godotenv.Load()

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
	conn, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		slog.Error("failed to connect to mongodb", "err", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Disconnect(context.Background()) }()
	mdb := conn.Database(cfg.MongoName)

	rds := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Username: cfg.RedisUser,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() { _ = rds.Close() }()

	rawStore, err := storageFactory.New(context.Background(), cfg, cfg.RawBucketName, false)
	if err != nil {
		slog.Error("failed to initialize raw storage", "err", err)
		os.Exit(1)
	}
	compressedStore, err := storageFactory.New(context.Background(), cfg, cfg.CompressedBucketName, true)
	if err != nil {
		slog.Error("failed to initialize compressed storage", "err", err)
		os.Exit(1)
	}

	processor := service.NewAttachmentProcessorService(
		mongoDatabase.NewPackDraftRepository(mdb),
		mongoDatabase.NewPackRepository(mdb),
		storage.Raw{Storage: rawStore},
		storage.Compressed{Storage: compressedStore},
		pubsub.NewChannelGetter(rds),
		redisCache.NewAttemptCounter(rds),
	)

	// Its own client/pool: a blocking queue consumer can hold connections
	// checked out for a long time, which would starve the attempt-counter
	// calls above if they shared one pool.
	queueRds := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Username: cfg.RedisUser,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() { _ = queueRds.Close() }()

	consumer, err := queueFactory.New(context.Background(), cfg, queueRds, maxConcurrentProcessors)
	if err != nil {
		slog.Error("failed to initialize notification queue consumer", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	servAddress := cfg.Host + ":" + cfg.Port
	srv := &http.Server{Addr: servAddress, Handler: mux}

	go func() {
		slog.Info("starting attachment processor", "addr", servAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		if err := consumer.Consume(ctx, processor.Process); err != nil {
			slog.Error("notification consumer stopped unexpectedly", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down attachment processor")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown gracefully", "err", err)
		os.Exit(1)
	}

	select {
	case <-consumerDone:
	case <-shutdownCtx.Done():
		slog.Error("notification consumer did not drain in time")
	}

	slog.Info("attachment processor shutdown gracefully")
}
