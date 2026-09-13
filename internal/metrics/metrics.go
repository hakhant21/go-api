package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "route", "status"},
	)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_hits_total", Help: "Cache hits."},
		[]string{"cache"},
	)
	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_misses_total", Help: "Cache misses."},
		[]string{"cache"},
	)
	SingleflightShared = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "singleflight_shared_total", Help: "Shared calls."},
		[]string{"key_prefix"},
	)
	SingleflightLoads = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "singleflight_loads_total", Help: "Loads triggered."},
		[]string{"key_prefix"},
	)
)
