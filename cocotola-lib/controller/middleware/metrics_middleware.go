package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics holds Prometheus metrics collectors for HTTP request instrumentation.
type PrometheusMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewPrometheusMetrics creates and registers HTTP metrics with the given registerer.
func NewPrometheusMetrics(reg prometheus.Registerer) *PrometheusMetrics {
	m := &PrometheusMetrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{ //nolint:exhaustruct
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{ //nolint:exhaustruct
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
	reg.MustRegister(m.requestsTotal, m.requestDuration)

	return m
}

// Middleware returns a Gin middleware that collects HTTP metrics for Prometheus.
func (m *PrometheusMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		m.requestsTotal.WithLabelValues(method, path, status).Inc()
		m.requestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

var (
	defaultMetrics     *PrometheusMetrics //nolint:gochecknoglobals
	defaultMetricsOnce sync.Once          //nolint:gochecknoglobals
)

// PrometheusMiddleware returns a Gin middleware that collects HTTP metrics using the default Prometheus registerer.
func PrometheusMiddleware() gin.HandlerFunc {
	defaultMetricsOnce.Do(func() {
		defaultMetrics = NewPrometheusMetrics(prometheus.DefaultRegisterer)
	})

	return defaultMetrics.Middleware()
}
