package server

import (
	"context"
	"fmt"
	"time"

	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/metrics"
)

// FlushEventLog copies the room's buffered events to the database; safe to call repeatedly.
func FlushEventLog(ctx context.Context, roomCache cache.Room, eventRepository repository.GameEvent, roomId string) error {
	start := time.Now()
	err := flushEventLog(ctx, roomCache, eventRepository, roomId)
	metrics.GameLogFlushDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.GameLogFlushErrorsTotal.Inc()
	}
	return err
}

func flushEventLog(ctx context.Context, roomCache cache.Room, eventRepository repository.GameEvent, roomId string) error {
	events, err := roomCache.GetEvents(ctx, roomId)
	if err != nil {
		return fmt.Errorf("reading event log: %w", err)
	}
	if err := eventRepository.InsertMany(ctx, roomId, events); err != nil {
		return fmt.Errorf("storing event log: %w", err)
	}
	return nil
}
