package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/middleware"
	adminsvc "github.com/hakhant21/go-starter/internal/service/admin"
)

type AdminHandler struct{ svc *adminsvc.Service }

func NewAdminHandler(svc *adminsvc.Service) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/admin")
	g.GET("/users", h.listUsers)
	g.GET("/users/:id", h.getUser)
	g.PATCH("/users/:id/active", h.setUserActive)
	g.POST("/users/:id/roles", h.assignRole)
	g.DELETE("/users/:id/roles/:role", h.revokeRole)

	g.GET("/roles", h.listRoles)
	g.POST("/roles", h.createRole)
	g.PUT("/roles/:name/permissions", h.setRolePermissions)

	g.GET("/permissions", h.listPermissions)
	g.POST("/permissions", h.createPermission)
}

// @Summary List users for administration
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search term"
// @Param role query string false "Role name"
// @Param active query bool false "Active status"
// @Success 200 {object} PaginatedResponse
// @Router /admin/users [get]
func (h *AdminHandler) listUsers(c *gin.Context) {
	var q dto.ListAdminUsersQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	users, total, err := h.svc.ListUsers(c.Request.Context(), q)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: users, Total: total, Page: q.Page, Limit: q.Limit})
}

// @Summary Get user for administration
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.AdminUserResponse
// @Router /admin/users/{id} [get]
func (h *AdminHandler) getUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := h.svc.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// @Summary Set user active status
// @Tags admin
// @Accept json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.SetActiveRequest true "Active status"
// @Success 204
// @Router /admin/users/{id}/active [patch]
func (h *AdminHandler) setUserActive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.svc.SetUserActive(c.Request.Context(), uint(id), req.Active); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Assign role to user
// @Tags admin
// @Accept json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.AssignRoleRequest true "Role request"
// @Success 204
// @Router /admin/users/{id}/roles [post]
func (h *AdminHandler) assignRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	if err := h.svc.AssignRole(c.Request.Context(), uint(id), req.RoleName); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Revoke role from user
// @Tags admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param role path string true "Role name"
// @Success 204
// @Router /admin/users/{id}/roles/{role} [delete]
func (h *AdminHandler) revokeRole(c *gin.Context) {
	actorID, _ := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	roleName := c.Param("role")
	if err := h.svc.RevokeRole(c.Request.Context(), actorID, uint(id), roleName); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary List roles
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.RoleResponse
// @Router /admin/roles [get]
func (h *AdminHandler) listRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// @Summary Create role
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRoleRequest true "Role request"
// @Success 201 {object} dto.RoleResponse
// @Router /admin/roles [post]
func (h *AdminHandler) createRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, role)
}

// @Summary Set role permissions
// @Tags admin
// @Accept json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Param request body dto.SetRolePermissionsRequest true "Permissions request"
// @Success 204
// @Router /admin/roles/{name}/permissions [put]
func (h *AdminHandler) setRolePermissions(c *gin.Context) {
	roleName := c.Param("name")
	var req dto.SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	if err := h.svc.SetRolePermissions(c.Request.Context(), roleName, req.Permissions); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary List permissions
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.PermissionResponse
// @Router /admin/permissions [get]
func (h *AdminHandler) listPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, perms)
}

// @Summary Create permission
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePermissionRequest true "Permission request"
// @Success 201 {object} dto.PermissionResponse
// @Router /admin/permissions [post]
func (h *AdminHandler) createPermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	p, err := h.svc.CreatePermission(c.Request.Context(), req)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}
