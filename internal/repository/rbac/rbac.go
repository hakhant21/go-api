package rbac

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hakhant21/go-starter/internal/model"
)

type Repository interface {
	AssignRoleByName(ctx context.Context, userID uint, roleName string) error
	RevokeRoleByName(ctx context.Context, userID uint, roleName string) error
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	GetUserRoleNames(ctx context.Context, userID uint) ([]string, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	CreateRole(ctx context.Context, role *model.Role) error
	SetRolePermissionsByName(ctx context.Context, roleName string, permNames []string) ([]uint, error)
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	CreatePermission(ctx context.Context, p *model.Permission) error
}

type repo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) AssignRoleByName(ctx context.Context, userID uint, roleName string) error {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("role %q not found: %w", roleName, err)
	}
	user := model.User{ID: userID}
	return r.db.WithContext(ctx).Model(&user).Association("Roles").Append(&role)
}

func (r *repo) RevokeRoleByName(ctx context.Context, userID uint, roleName string) error {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("role %q not found: %w", roleName, err)
	}
	user := model.User{ID: userID}
	return r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(&role)
}

func (r *repo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var perms []string
	err := r.db.WithContext(ctx).
		Table("permissions p").
		Select("DISTINCT p.name").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Pluck("p.name", &perms).Error
	return perms, err
}

func (r *repo) GetUserRoleNames(ctx context.Context, userID uint) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("roles r").
		Select("r.name").
		Joins("JOIN user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Pluck("r.name", &names).Error
	return names, err
}

func (r *repo) ListRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *repo) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repo) SetRolePermissionsByName(ctx context.Context, roleName string, permNames []string) ([]uint, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role %q not found: %w", roleName, err)
	}
	var perms []model.Permission
	if len(permNames) > 0 {
		if err := r.db.WithContext(ctx).Where("name IN ?", permNames).Find(&perms).Error; err != nil {
			return nil, err
		}
	}
	if err := r.db.WithContext(ctx).Model(&role).Association("Permissions").Replace(perms); err != nil {
		return nil, err
	}
	var userIDs []uint
	if err := r.db.WithContext(ctx).Table("user_roles").
		Where("role_id = ?", role.ID).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *repo) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.WithContext(ctx).Order("name").Find(&perms).Error
	return perms, err
}

func (r *repo) CreatePermission(ctx context.Context, p *model.Permission) error {
	return r.db.WithContext(ctx).Create(p).Error
}
