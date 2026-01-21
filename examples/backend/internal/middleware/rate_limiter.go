package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter はIPアドレスごとのレート制限を管理
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
}

// NewRateLimiter は新しいレート制限インスタンスを作成
// limit: 1秒あたりのリクエスト数（100req/min = 100/60 = 約1.67req/sec）
// burst: バーストサイズ（短時間に許可される最大リクエスト数）
func NewRateLimiter(reqPerMin int) *RateLimiter {
	limit := rate.Limit(float64(reqPerMin) / 60.0)
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    reqPerMin / 6, // バーストサイズは10秒分のリクエスト数
	}
}

// getVisitor はIPアドレスに対応するリミッターを取得または作成
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// RateLimiterMiddleware はレート制限ミドルウェアを返す
func RateLimiterMiddleware(reqPerMin int) gin.HandlerFunc {
	limiter := NewRateLimiter(reqPerMin)

	return func(c *gin.Context) {
		// クライアントのIPアドレスを取得
		ip := c.ClientIP()

		// IPアドレスに対応するリミッターを取得
		visitor := limiter.getVisitor(ip)

		// レート制限をチェック
		if !visitor.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
