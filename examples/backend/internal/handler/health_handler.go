package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler はヘルスチェック関連のHTTPリクエストを処理
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler はHealthHandlerの新しいインスタンスを作成
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthCheck はアプリケーションとデータベースのヘルスチェックを実行
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "ok"
	dbStatus := "ok"

	// データベース接続のチェック（2秒タイムアウト）
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// シンプルなpingクエリでデータベース接続を確認
	err := h.db.PingContext(ctx)
	if err != nil {
		dbStatus = "error"
		status = "degraded"
	}

	// ヘルスチェックレスポンス
	response := gin.H{
		"status": status,
		"timestamp": time.Now().Format(time.RFC3339),
		"services": gin.H{
			"database": dbStatus,
		},
	}

	// ステータスに応じてHTTPステータスコードを設定
	httpStatus := http.StatusOK
	if status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response)
}
