// @title           Starter API
// @version         1.0
// @description     Gin + GORM REST API starter.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hakhant21/go-starter/internal/cache"
	"github.com/hakhant21/go-starter/internal/config"
	"github.com/hakhant21/go-starter/internal/database"
	"github.com/hakhant21/go-starter/internal/email"
	"github.com/hakhant21/go-starter/internal/handler"
	refreshtokenrepo "github.com/hakhant21/go-starter/internal/repository/refresh_token"
	rbacrepo "github.com/hakhant21/go-starter/internal/repository/rbac"
	tokenrepo "github.com/hakhant21/go-starter/internal/repository/token"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	adminsvc "github.com/hakhant21/go-starter/internal/service/admin"
	authsvc "github.com/hakhant21/go-starter/internal/service/auth"
	cleanupsvc "github.com/hakhant21/go-starter/internal/service/cleanup"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
	usersvc "github.com/hakhant21/go-starter/internal/service/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.New(cfg.DatabaseURL, cfg.IsProduction())
	if err != nil {
		return err
	}
	if err := database.AutoMigrate(db); err != nil {
		return err
	}
	if err := database.SeedRBAC(db); err != nil {
		return err
	}

	rdb, err := cache.NewRedis(cfg.RedisURL)
	if err != nil {
		return err
	}

	var mailer email.Sender
	if cfg.IsProduction() && cfg.SMTPHost != "" {
		mailer = email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	} else {
		mailer = email.NewConsoleSender()
	}

	userRepo := userrepo.New(db)
	refreshRepo := refreshtokenrepo.New(db)
	tokenRepo := tokenrepo.New(db)
	rbacRepo := rbacrepo.New(db)

	rbacCache := cache.NewRBACCache(rdb)
	userCache := cache.NewUserCache(rdb)
	lock := cache.NewLock(rdb)

	rbacSvc := rbacsvc.New(rbacRepo, rbacCache, lock)
	authSvc := authsvc.New(userRepo, refreshRepo, tokenRepo, mailer,
		cfg.JWTSecret, int64(cfg.JWTExpiry.Seconds()), cfg.RefreshTTL, cfg.AppBaseURL)
	userSvc := usersvc.New(userRepo, authSvc, rbacSvc, userCache)
	adminSvc := adminsvc.New(userRepo, rbacRepo, rbacSvc)
	cleanupSvc := cleanupsvc.New(tokenRepo, refreshRepo)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	go cleanupSvc.Run(rootCtx, time.Hour)

	if cfg.BootstrapAdmin != "" {
		if u, err := userRepo.GetByEmail(rootCtx, cfg.BootstrapAdmin); err == nil {
			_ = rbacSvc.AssignRole(rootCtx, u.ID, "admin")
			slog.Info("bootstrap admin assigned", "email", cfg.BootstrapAdmin)
		}
	}

	healthHandler := handler.NewHealthHandler(db, rdb)

	router := handler.NewRouter(handler.Deps{
		UserSvc:        userSvc,
		AuthSvc:        authSvc,
		RBACSvc:        rbacSvc,
		AdminSvc:       adminSvc,
		Redis:          rdb,
		Secret:         cfg.JWTSecret,
		IsProd:         cfg.IsProduction(),
		AllowedOrigins: cfg.AllowedOrigins,
		HealthHandler:  healthHandler,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
