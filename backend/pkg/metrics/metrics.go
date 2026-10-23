package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sgame_http_requests_total",
		Help: "Total HTTP requests by method, path and status",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sgame_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	WSConnectionsActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sgame_ws_connections_active",
		Help: "Currently active WebSocket connections",
	}, []string{"type"})

	WSConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sgame_ws_connections_total",
		Help: "Total WebSocket connections established",
	}, []string{"type"})

	RoomsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sgame_rooms_active",
		Help: "Number of active game rooms",
	})

	GameEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sgame_game_events_total",
		Help: "Total game events processed",
	}, []string{"event", "direction"})

	GameEventsLogged = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sgame_game_events_logged_total",
		Help: "Game log events appended to the room log by type",
	}, []string{"type"})

	GameLogFlushDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "sgame_game_log_flush_duration_seconds",
		Help:    "Time spent persisting a room's game log to the database",
		Buckets: prometheus.DefBuckets,
	})

	GameLogFlushErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sgame_game_log_flush_errors_total",
		Help: "Failed attempts to persist a room's game log",
	})

	AttachmentsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sgame_attachments_processed_total",
		Help: "Total attachment notifications processed by outcome (ready, failed, skipped, dropped, retry)",
	}, []string{"outcome"})

	AttachmentProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "sgame_attachment_processing_duration_seconds",
		Help:    "Time spent processing a single attachment notification",
		Buckets: prometheus.DefBuckets,
	})

	AttachmentsProcessingInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sgame_attachments_processing_in_flight",
		Help: "Number of attachment notifications currently being processed",
	})
)
