package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hakhant21/go-starter/pkg/jwt"
)

type ctxKey string

const (
	UserIDKey ctxKey = "userID"
	EmailKey  ctxKey = "email"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		claims, err := jwt.Parse(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(EmailKey), claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get(string(UserIDKey))
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

func GetEmail(c *gin.Context) string {
	v, _ := c.Get(string(EmailKey))
	s, _ := v.(string)
	return s
}
