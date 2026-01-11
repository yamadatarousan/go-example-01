package middleware

import (
	"net/http"

	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware は管理者権限をチェックするミドルウェアを返す
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		appClaims, ok := claims.(*service.AppClaims)
		if !ok || appClaims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		c.Next()
	}
}
