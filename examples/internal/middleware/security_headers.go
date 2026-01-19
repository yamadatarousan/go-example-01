package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware はセキュリティヘッダーを追加するミドルウェア
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS攻撃対策: ブラウザにコンテンツタイプのスニッフィングを防止させる
		c.Header("X-Content-Type-Options", "nosniff")

		// クリックジャッキング対策: iframeでの表示を禁止
		c.Header("X-Frame-Options", "DENY")

		// XSS攻撃対策: ブラウザのXSSフィルターを有効化
		c.Header("X-XSS-Protection", "1; mode=block")

		// コンテンツセキュリティポリシー: 自サイトのコンテンツのみ許可
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Referrerポリシー: リファラー情報の送信を制限
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: 不要な機能を無効化
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HTTPS接続時のみHSTSヘッダーを追加
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
