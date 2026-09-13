package dto

type AdminUserResponse struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Active        bool     `json:"active"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"roles"`
	Permissions   []string `json:"permissions"`
	CreatedAt     string   `json:"created_at"`
}

type ListAdminUsersQuery struct {
	Page     int    `form:"page"   validate:"omitempty,min=1"`
	Limit    int    `form:"limit"  validate:"omitempty,min=1,max=100"`
	Search   string `form:"search" validate:"omitempty,max=100"`
	RoleName string `form:"role"   validate:"omitempty,max=50"`
	Active   *bool  `form:"active"`
}

type AssignRoleRequest struct {
	RoleName string `json:"role_name" validate:"required,max=50"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name"        validate:"required,min=2,max=50"`
	Description string   `json:"description" validate:"omitempty,max=255"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,max=100"`
}

type SetRolePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required,dive,max=100"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type RoleResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SetActiveRequest struct {
	Active bool `json:"active"`
}
