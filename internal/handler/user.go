package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hakhant21/go-starter/internal/dto"
	usersvc "github.com/hakhant21/go-starter/internal/service/user"
)

type UserHandler struct{ svc *usersvc.Service }

func NewUserHandler(svc *usersvc.Service) *UserHandler { return &UserHandler{svc: svc} }

func (h *UserHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/users")
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *UserHandler) create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	u, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.UserToResponse(u))
}

func (h *UserHandler) list(c *gin.Context) {
	var q dto.ListUsersQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	users, total, err := h.svc.List(c.Request.Context(), q.Page, q.Limit)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	out := make([]*dto.UserResponse, 0, len(users))
	for i := range users {
		out = append(out, dto.UserToResponse(&users[i]))
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: out, Total: total, Page: q.Page, Limit: q.Limit})
}

func (h *UserHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.UserToResponse(u))
}

func (h *UserHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	u, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.UserToResponse(u))
}

func (h *UserHandler) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
