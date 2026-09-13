package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hakhant21/go-starter/internal/middleware"
	"github.com/hakhant21/go-starter/internal/rbac"
	adminsvc "github.com/hakhant21/go-starter/internal/service/admin"
	authsvc "github.com/hakhant21/go-starter/internal/service/auth"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
	usersvc "github.com/hakhant21/go-starter/internal/service/user"
)

type Deps struct {
	UserSvc        *usersvc.Service
	AuthSvc        *authsvc.Service
	RBACSvc        *rbacsvc.Service
	AdminSvc       *adminsvc.Service
	Redis          *redis.Client
	Secret         string
	IsProd         bool
	AllowedOrigins []string
	HealthHandler  *HealthHandler
}

func NewRouter(d Deps) *gin.Engine {
	if d.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.Metrics())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(d.AllowedOrigins))
	r.Use(middleware.BodyLimit(1 << 20))
	r.Use(middleware.RateLimit(d.Redis, "global", 300, time.Minute))

	if d.HealthHandler != nil {
		d.HealthHandler.Register(r)
	}

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authMiddleware := middleware.Auth(d.Secret)

	v1 := r.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.RateLimit(d.Redis, "auth", 10, time.Minute))
	authHandler := NewAuthHandler(d.AuthSvc, d.UserSvc, d.RBACSvc)
	authHandler.Register(authGroup)

	v1.GET("/auth/me", authMiddleware, authHandler.Me)

	NewUserHandler(d.UserSvc).Register(v1)

	adminGroup := v1.Group("/admin")
	adminGroup.Use(authMiddleware, middleware.RequirePermission(d.RBACSvc, rbac.PermRolesManage))
	NewAdminHandler(d.AdminSvc).Register(adminGroup)

	return r
}
