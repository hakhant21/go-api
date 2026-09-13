package admin

import (
	"context"
	"time"

	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
	rbacrepo "github.com/hakhant21/go-starter/internal/repository/rbac"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	"github.com/hakhant21/go-starter/internal/service"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
)

type Service struct {
	userRepo userrepo.Repository
	rbacRepo rbacrepo.Repository
	rbacSvc  *rbacsvc.Service
}

func New(u userrepo.Repository, r rbacrepo.Repository, rs *rbacsvc.Service) *Service {
	return &Service{userRepo: u, rbacRepo: r, rbacSvc: rs}
}

func (s *Service) ListUsers(ctx context.Context, q dto.ListAdminUsersQuery) ([]dto.AdminUserResponse, int64, error) {
	page, limit := q.Page, q.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := s.userRepo.ListFiltered(ctx, userrepo.Filter{
		Search: q.Search, RoleName: q.RoleName, Active: q.Active,
		Limit: limit, Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AdminUserResponse, 0, len(users))
	for i := range users {
		u := &users[i]
		roles, _ := s.rbacRepo.GetUserRoleNames(ctx, u.ID)
		perms, _ := s.rbacSvc.Permissions(ctx, u.ID)
		out = append(out, dto.AdminUserResponse{
			ID: u.ID, Name: u.Name, Email: u.Email,
			Active: u.Active, EmailVerified: u.EmailVerified,
			Roles: roles, Permissions: perms,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, total, nil
}

func (s *Service) GetUser(ctx context.Context, id uint) (*dto.AdminUserResponse, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	roles, _ := s.rbacRepo.GetUserRoleNames(ctx, u.ID)
	perms, _ := s.rbacSvc.Permissions(ctx, u.ID)
	return &dto.AdminUserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email,
		Active: u.Active, EmailVerified: u.EmailVerified,
		Roles: roles, Permissions: perms,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) SetUserActive(ctx context.Context, id uint, active bool) error {
	if err := s.userRepo.SetActive(ctx, id, active); err != nil {
		return err
	}
	return s.rbacSvc.CacheInvalidate(ctx, id)
}

func (s *Service) AssignRole(ctx context.Context, userID uint, roleName string) error {
	return s.rbacSvc.AssignRole(ctx, userID, roleName)
}

func (s *Service) RevokeRole(ctx context.Context, actorID, userID uint, roleName string) error {
	if actorID == userID && roleName == "admin" {
		return service.ErrForbidden
	}
	return s.rbacSvc.RevokeRole(ctx, userID, roleName)
}

func (s *Service) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.rbacRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RoleResponse, 0, len(roles))
	for _, r := range roles {
		perms := make([]string, 0, len(r.Permissions))
		for _, p := range r.Permissions {
			perms = append(perms, p.Name)
		}
		out = append(out, dto.RoleResponse{ID: r.ID, Name: r.Name, Description: r.Description, Permissions: perms})
	}
	return out, nil
}

func (s *Service) CreateRole(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := &model.Role{Name: req.Name, Description: req.Description}
	if err := s.rbacRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	if len(req.Permissions) > 0 {
		if err := s.rbacSvc.SetRolePermissions(ctx, role.Name, req.Permissions); err != nil {
			return nil, err
		}
	}
	return s.roleResponse(ctx, role.Name)
}

func (s *Service) SetRolePermissions(ctx context.Context, roleName string, perms []string) error {
	return s.rbacSvc.SetRolePermissions(ctx, roleName, perms)
}

func (s *Service) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := s.rbacRepo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		out = append(out, dto.PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description})
	}
	return out, nil
}

func (s *Service) CreatePermission(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	p := &model.Permission{Name: req.Name, Description: req.Description}
	if err := s.rbacRepo.CreatePermission(ctx, p); err != nil {
		return nil, err
	}
	return &dto.PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description}, nil
}

func (s *Service) roleResponse(ctx context.Context, name string) (*dto.RoleResponse, error) {
	roles, err := s.rbacRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.Name == name {
			perms := make([]string, 0, len(r.Permissions))
			for _, p := range r.Permissions {
				perms = append(perms, p.Name)
			}
			return &dto.RoleResponse{ID: r.ID, Name: r.Name, Description: r.Description, Permissions: perms}, nil
		}
	}
	return nil, repository.ErrNotFound
}
