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

func (h *AdminHandler) listRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

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

func (h *AdminHandler) listPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, perms)
}

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
