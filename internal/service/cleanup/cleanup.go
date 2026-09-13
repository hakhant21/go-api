package cleanup

import (
	"context"
	"log/slog"
	"time"

	refreshtokenrepo "github.com/hakhant21/go-starter/internal/repository/refresh_token"
	tokenrepo "github.com/hakhant21/go-starter/internal/repository/token"
)

type Service struct {
	tokenRepo   tokenrepo.Repository
	refreshRepo refreshtokenrepo.Repository
}

func New(t tokenrepo.Repository, r refreshtokenrepo.Repository) *Service {
	return &Service{tokenRepo: t, refreshRepo: r}
}

func (s *Service) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.tokenRepo.DeleteExpired(ctx); err != nil {
				slog.Warn("cleanup tokens", "err", err)
			}
			if err := s.refreshRepo.DeleteExpired(ctx); err != nil {
				slog.Warn("cleanup refresh", "err", err)
			}
		}
	}
}
