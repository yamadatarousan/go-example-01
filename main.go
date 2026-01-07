package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Todo struct {
	ID     int    `json:"id"`
	Name   string `json:"name" binding:"required"`
	UserID int    `json:"user_id"`
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

type TodoHandler struct {
	repo *TodoRepository
}

func NewTodoHandler(repo *TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

func (h *TodoHandler) getTodos(c *gin.Context) error {
	todos, err := h.repo.FindAll(1)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, todos)
	return nil
}

func (h *TodoHandler) createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		return err
	}
	todo, err := h.repo.Create(newTodo)
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, todo)
	return nil
}

type AppHandler func(c *gin.Context) error

func errorHandler(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			log.Printf("Error occurred: %v", err)
			// その他の予期せぬエラーの場合
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
			})
			return
		}
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
	db, err = sql.Open("pgx", dsn)
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

	// --- 依存関係の構築 (DI: Dependency Injection) ---
	// 1. リポジトリのインスタンスを作成
	repo := NewTodoRepository(db)
	// 2. ハンドラのインスタンスを作成し、リポジトリを注入
	todoHandler := NewTodoHandler(repo)

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
	router.GET("todos", errorHandler(todoHandler.getTodos))
	router.POST("todos", errorHandler(todoHandler.createTodo))
	router.Run()
}
