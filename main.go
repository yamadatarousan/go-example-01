package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func main() {
	router := gin.New()
	// カスタムログフォーマッタを定義します。
	logFromatter := func(param gin.LogFormatterParams) string {
		requestID := param.Keys["RequestID"]
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format(time.RFC3339), // timeパッケージの定数を明示的に使用
			requestID,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
		)
	}
	// ミドルウェアを .Use() で適用します。適用した順に実行されます。
	// 1. Recovery: panicが発生してもサーバーが落ちないようにする。
	router.Use(gin.Recovery())
	// 2. RequestID: これ以降の処理（ロガーなど）で使えるようにIDを生成する。
	router.Use(requestIDMiddleware())
	// 3. Logger: カスタムフォーマットのロガーを適用する。
	router.Use(gin.LoggerWithFormatter(logFromatter))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run()
}
