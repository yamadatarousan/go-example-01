package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-quickstart/examples/internal/config"
	"gin-quickstart/examples/internal/handler"
	"gin-quickstart/examples/internal/middleware"
	"gin-quickstart/examples/internal/repository"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// 設定の読み込み
	cfg := config.Load()

	// データベース接続の初期化
	db, err := initDB(cfg.Database)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// 依存関係の構築 (Dependency Injection)
	// Repository層
	todoRepo := repository.NewTodoRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Service層
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	todoService := service.NewTodoService(todoRepo)
	adminService := service.NewAdminService(userRepo)

	// Handler層
	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	adminHandler := handler.NewAdminHandler(adminService)

	// ルーターの設定
	router := setupRouter(cfg, authService, userHandler, todoHandler, adminHandler)

	// Graceful Shutdownの実装
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// サーバーをゴルーチンで起動（非同期処理）
	go func() {
		log.Printf("サーバーをポート%sで起動中...", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("サーバーの起動に失敗しました: %v", err)
		}
	}()

	// 終了シグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウン中...")

	// 5秒以内に既存のリクエストの処理を終了
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("サーバーの強制終了: %v", err)
	}

	log.Println("サーバーが終了しました")
}

// initDB はデータベース接続を初期化
func initDB(dbConfig config.DatabaseConfig) (*sql.DB, error) {
	dsn := dbConfig.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベースのオープンに失敗: %w", err)
	}

	// 接続確認
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("データベースへの接続に失敗: %w", err)
	}

	log.Println("データベースに正常に接続しました")
	return db, nil
}

// setupRouter はGinルーターを設定
func setupRouter(
	cfg *config.Config,
	authService *service.AuthService,
	userHandler *handler.UserHandler,
	todoHandler *handler.TodoHandler,
	adminHandler *handler.AdminHandler,
) *gin.Engine {
	router := gin.New()

	// CORSの設定
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.AllowOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Authorization", "Content-Type"}
	router.Use(cors.New(corsConfig))

	// カスタムログフォーマッタの定義
	logFormatter := func(param gin.LogFormatterParams) string {
		requestID := param.Keys["RequestID"]
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format(time.RFC3339),
			requestID,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
		)
	}

	// ミドルウェアの適用（適用した順に実行）
	router.Use(gin.Recovery())                        // panicからの回復
	router.Use(middleware.RequestIDMiddleware())      // リクエストIDの生成
	router.Use(gin.LoggerWithFormatter(logFormatter)) // ログ出力

	// ヘルスチェック
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 認証エンドポイント
	router.POST("/signup", handler.ErrorHandler(userHandler.Signup))
	router.POST("/login", handler.ErrorHandler(userHandler.Login))

	// 認証が必要なエンドポイント
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(authService))
	{
		// TODOエンドポイント
		v1.GET("/todos", handler.ErrorHandler(todoHandler.GetTodos))
		v1.GET("/todos/:id", handler.ErrorHandler(todoHandler.GetTodo))
		v1.POST("/todos", handler.ErrorHandler(todoHandler.CreateTodo))
		v1.PUT("/todos/:id", handler.ErrorHandler(todoHandler.UpdateTodo))
		v1.DELETE("/todos/:id", handler.ErrorHandler(todoHandler.DeleteTodo))

		// 管理者専用エンドポイント
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.AdminMiddleware())
		{
			adminRoutes.GET("/users", handler.ErrorHandler(adminHandler.GetAllUsers))
		}
	}

	return router
}
