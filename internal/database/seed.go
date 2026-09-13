package database

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/rbac"
)

func SeedRBAC(db *gorm.DB) error {
	permSet := map[string]struct{}{}
	for _, perms := range rbac.DefaultRoles {
		for _, p := range perms {
			permSet[p] = struct{}{}
		}
	}

	for name := range permSet {
		var existing model.Permission
		if err := db.Where("name = ?", name).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&model.Permission{Name: name}).Error; err != nil {
			return err
		}
	}

	for roleName, permNames := range rbac.DefaultRoles {
		var role model.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			role = model.Role{Name: roleName, Description: roleName + " role"}
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		}
		var perms []model.Permission
		if err := db.Where("name IN ?", permNames).Find(&perms).Error; err != nil {
			return err
		}
		if err := db.Model(&role).Association("Permissions").Replace(perms); err != nil {
			return err
		}
	}

	slog.Info("RBAC seeded")
	return nil
}
