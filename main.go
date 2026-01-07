package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func getTodoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		todos := []string{"Buy groceries", "Walk the dog", "Read a book"}
		c.JSON(200, gin.H{
			"todos": todos,
		})
	}
}

func postTodoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		todos := []string{"Buy groceries", "Walk the dog", "Read a book"}
		c.JSON(200, gin.H{
			"todos": todos,
		})
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var db *sql.DB

func initDB() {
	var err error
	// --- PostgreSQLへの接続情報 (DSN: Data Source Name) ---
	// 環境変数から接続情報を取得（Docker環境対応）
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5435")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "todo_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// 接続確認
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the database")
}

func main() {
	initDB()
	// Ginのモードを設定します。デフォルトは "debug" モードです。
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
	router.GET("todos", getTodoHandler())
	router.POST("todos", postTodoHandler())
	router.Run()
}
