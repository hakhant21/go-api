package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/hakhant21/go-starter/internal/repository"
	"github.com/hakhant21/go-starter/internal/service"
)

var validate = validator.New()

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Data  any   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

func respondValidationError(c *gin.Context, err error) {
	details := map[string]string{}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[fe.Field()] = validationMessage(fe)
		}
	}
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: details})
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	default:
		return "invalid"
	}
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, "resource not found")
	case errors.Is(err, service.ErrInvalidInput):
		respondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		respondError(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		respondError(c, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, service.ErrForbidden):
		respondError(c, http.StatusForbidden, "forbidden")
	case errors.Is(err, service.ErrEmailNotVerified):
		respondError(c, http.StatusForbidden, "email not verified")
	case errors.Is(err, service.ErrTokenInvalid):
		respondError(c, http.StatusBadRequest, "invalid or expired token")
	case errors.Is(err, service.ErrEmailAlreadyVerified):
		respondError(c, http.StatusConflict, "email already verified")
	default:
		slog.Error("unhandled error", "err", err)
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}
