package middleware

import (
	"log"
	"net/http"
	"strings"

	"gin-quickstart/examples/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware はJWT認証を行うミドルウェアを返す
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// "Bearer <token>" という形式を期待
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		claims, err := authService.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		c.Set("claims", claims)
		log.Println("認証済みユーザーID:", claims.Subject, "ロール:", claims.Role)
		c.Next()
	}
}
