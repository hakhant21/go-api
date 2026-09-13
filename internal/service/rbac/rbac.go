package rbac

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/hakhant21/go-starter/internal/cache"
	"github.com/hakhant21/go-starter/internal/model"
	rbacrepo "github.com/hakhant21/go-starter/internal/repository/rbac"
)

type Service struct {
	repo  rbacrepo.Repository
	cache *cache.RBACCache
	lock  *cache.Lock
	group singleflight.Group
}

func New(repo rbacrepo.Repository, c *cache.RBACCache, l *cache.Lock) *Service {
	return &Service{repo: repo, cache: c, lock: l}
}

func (s *Service) Permissions(ctx context.Context, userID uint) ([]string, error) {
	return cache.GetOrLoad(
		ctx,
		&s.group,
		fmt.Sprintf("perms:%d", userID),
		func(ctx context.Context) ([]string, bool) {
			return s.cache.Get(ctx, userID)
		},
		func(ctx context.Context, perms []string) {
			s.cache.Set(ctx, userID, perms)
		},
		func(ctx context.Context) ([]string, error) {
			return s.repo.GetUserPermissions(ctx, userID)
		},
	)
}

func (s *Service) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	perms, err := s.Permissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) HasAnyPermission(ctx context.Context, userID uint, wants ...string) (bool, error) {
	perms, err := s.Permissions(ctx, userID)
	if err != nil {
		return false, err
	}
	set := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		set[p] = struct{}{}
	}
	for _, w := range wants {
		if _, ok := set[w]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) List(ctx context.Context) ([]model.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) AssignRole(ctx context.Context, userID uint, roleName string) error {
	if err := s.repo.AssignRoleByName(ctx, userID, roleName); err != nil {
		return err
	}
	s.cache.Invalidate(ctx, userID)
	return nil
}

func (s *Service) RevokeRole(ctx context.Context, userID uint, roleName string) error {
	if err := s.repo.RevokeRoleByName(ctx, userID, roleName); err != nil {
		return err
	}
	s.cache.Invalidate(ctx, userID)
	return nil
}

func (s *Service) SetRolePermissions(ctx context.Context, roleName string, perms []string) error {
	userIDs, err := s.repo.SetRolePermissionsByName(ctx, roleName, perms)
	if err != nil {
		return err
	}
	s.cache.InvalidateUsers(ctx, userIDs)
	return nil
}

func (s *Service) CacheInvalidate(ctx context.Context, userID uint) error {
	s.cache.Invalidate(ctx, userID)
	return nil
}
