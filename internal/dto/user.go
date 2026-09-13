package dto

import (
	"time"

	"github.com/hakhant21/go-starter/internal/model"
)

type CreateUserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"  validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserResponse struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Active        bool     `json:"active"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"roles,omitempty"`
	Permissions   []string `json:"permissions,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

type ListUsersQuery struct {
	Page  int `form:"page"  validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

func UserToResponse(u *model.User) *UserResponse {
	if u == nil {
		return nil
	}
	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Active:        u.Active,
		EmailVerified: u.EmailVerified,
		Roles:         roles,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}
