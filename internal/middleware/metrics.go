package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hakhant21/go-starter/internal/metrics"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		metrics.HTTPRequestsTotal.
			WithLabelValues(c.Request.Method, route, statusClass(c.Writer.Status())).Inc()
		metrics.HTTPRequestDuration.
			WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

func statusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}
