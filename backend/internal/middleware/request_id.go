package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware はリクエストIDを生成し、コンテキストに設定するミドルウェアを返す
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
