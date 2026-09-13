package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"

	"github.com/hakhant21/go-starter/internal/cache"
	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	"github.com/hakhant21/go-starter/internal/service"
	authsvc "github.com/hakhant21/go-starter/internal/service/auth"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
)

type Service struct {
	repo      userrepo.Repository
	auth      *authsvc.Service
	rbac      *rbacsvc.Service
	userCache *cache.UserCache
	group     singleflight.Group
}

func New(repo userrepo.Repository, a *authsvc.Service, r *rbacsvc.Service, uc *cache.UserCache) *Service {
	return &Service{repo: repo, auth: a, rbac: r, userCache: uc}
}

func (s *Service) Create(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, service.ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		Name:          strings.TrimSpace(req.Name),
		Email:         email,
		Password:      string(hash),
		Active:        true,
		EmailVerified: false,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	if err := s.rbac.AssignRole(ctx, u.ID, "user"); err != nil {
		slog.Warn("assign default role", "err", err, "user_id", u.ID)
	}
	if err := s.auth.SendVerificationEmail(ctx, u); err != nil {
		slog.Warn("send verification email", "err", err, "user_id", u.ID)
	}
	return u, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*model.User, error) {
	return cache.GetOrLoad(
		ctx,
		&s.group,
		fmt.Sprintf("user:%d", id),
		func(ctx context.Context) (*model.User, bool) {
			return s.userCache.Get(ctx, id)
		},
		func(ctx context.Context, u *model.User) {
			s.userCache.Set(ctx, u)
		},
		func(ctx context.Context) (*model.User, error) {
			return s.repo.GetByID(ctx, id)
		},
	)
}

func (s *Service) List(ctx context.Context, page, limit int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, (page-1)*limit)
}

func (s *Service) Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		u.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		newEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		if newEmail != u.Email {
			if existing, err := s.repo.GetByEmail(ctx, newEmail); err == nil && existing.ID != u.ID {
				return nil, service.ErrEmailTaken
			}
			u.Email = newEmail
		}
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	s.userCache.Invalidate(ctx, id)
	return u, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.userCache.Invalidate(ctx, id)
	return nil
}

