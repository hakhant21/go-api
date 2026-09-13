package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/middleware"
	authsvc "github.com/hakhant21/go-starter/internal/service/auth"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
	usersvc "github.com/hakhant21/go-starter/internal/service/user"
)

type AuthHandler struct {
	svc     *authsvc.Service
	userSvc *usersvc.Service
	rbacSvc *rbacsvc.Service
}

func NewAuthHandler(svc *authsvc.Service, us *usersvc.Service, rs *rbacsvc.Service) *AuthHandler {
	return &AuthHandler{svc: svc, userSvc: us, rbacSvc: rs}
}

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/auth")
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
	g.POST("/verify-email", h.verifyEmail)
	g.POST("/resend-verification", h.resendVerification)
	g.POST("/forgot-password", h.forgotPassword)
	g.POST("/reset-password", h.resetPassword)
}

// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh request"
// @Success 200 {object} dto.AuthResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Logout
// @Tags auth
// @Accept json
// @Param request body dto.RefreshRequest true "Refresh request"
// @Success 204
// @Router /auth/logout [post]
func (h *AuthHandler) logout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Verify email
// @Tags auth
// @Accept json
// @Param request body dto.VerifyEmailRequest true "Verification request"
// @Success 204
// @Router /auth/verify-email [post]
func (h *AuthHandler) verifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	if err := h.svc.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Resend verification email
// @Tags auth
// @Accept json
// @Param request body dto.ResendVerificationRequest true "Resend request"
// @Success 204
// @Router /auth/resend-verification [post]
func (h *AuthHandler) resendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	_ = h.svc.ResendVerification(c.Request.Context(), req.Email)
	c.Status(http.StatusNoContent)
}

// @Summary Request password reset
// @Tags auth
// @Accept json
// @Param request body dto.ForgotPasswordRequest true "Password reset request"
// @Success 204
// @Router /auth/forgot-password [post]
func (h *AuthHandler) forgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	_ = h.svc.ForgotPassword(c.Request.Context(), req.Email)
	c.Status(http.StatusNoContent)
}

// @Summary Reset password
// @Tags auth
// @Accept json
// @Param request body dto.ResetPasswordRequest true "Password reset request"
// @Success 204
// @Router /auth/reset-password [post]
func (h *AuthHandler) resetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Get current user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.userSvc.Get(c.Request.Context(), uid)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	perms, _ := h.rbacSvc.Permissions(c.Request.Context(), uid)
	resp := dto.UserToResponse(u)
	resp.Permissions = perms
	c.JSON(http.StatusOK, resp)
}
