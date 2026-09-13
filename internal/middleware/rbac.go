package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
)

func RequirePermission(rbac *rbacsvc.Service, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		has, err := rbac.HasPermission(c.Request.Context(), uid, permission)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if !has {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
