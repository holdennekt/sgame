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
)
