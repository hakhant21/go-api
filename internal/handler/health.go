package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/health", h.health)
	r.GET("/ready", h.ready)
}

func (h *HealthHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	resp := gin.H{"status": "ok", "checks": gin.H{}}
	checks := resp["checks"].(gin.H)
	healthy := true

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		checks["postgres"] = "fail"
		healthy = false
	} else {
		checks["postgres"] = "ok"
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = "fail (degraded)"
	} else {
		checks["redis"] = "ok"
	}

	if !healthy {
		resp["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}
