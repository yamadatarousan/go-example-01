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
	categoryRepo := repository.NewCategoryRepository(db)       // Phase 2で追加
	notificationRepo := repository.NewNotificationRepository(db) // Phase 4で追加
	reminderRepo := repository.NewReminderRepository(db)       // Phase 4で追加
	// tagRepo := repository.NewTagRepository(db)         // Phase 2で追加（Phase 3で使用）

	// Service層
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	todoService := service.NewTodoService(todoRepo)
	adminService := service.NewAdminService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)                // Phase 2で追加
	notificationService := service.NewNotificationService(notificationRepo)     // Phase 4で追加
	_ = service.NewReminderService(reminderRepo, notificationRepo, todoRepo) // Phase 4で追加（将来のバックグラウンドワーカー用）

	// Handler層
	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)       // Phase 2で追加
	notificationHandler := handler.NewNotificationHandler(notificationService) // Phase 4で追加

	// ルーターの設定
	router := setupRouter(cfg, authService, userHandler, todoHandler, adminHandler, categoryHandler, notificationHandler)

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
	categoryHandler *handler.CategoryHandler,       // Phase 2で追加
	notificationHandler *handler.NotificationHandler, // Phase 4で追加
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

		// Phase 2で追加されたTODOエンドポイント
		v1.POST("/todos/:id/complete", handler.ErrorHandler(todoHandler.CompleteTodo)) // TODO完了
		v1.POST("/todos/:id/reopen", handler.ErrorHandler(todoHandler.ReopenTodo))     // TODO再開
		v1.GET("/todos/overdue", handler.ErrorHandler(todoHandler.GetOverdueTodos))    // 期限切れTODO
		v1.GET("/todos/today", handler.ErrorHandler(todoHandler.GetTodayTodos))        // 今日のTODO
		v1.GET("/todos/week", handler.ErrorHandler(todoHandler.GetThisWeekTodos))      // 今週のTODO

		// Phase 3で追加されたTODOエンドポイント
		v1.GET("/todos/search", handler.ErrorHandler(todoHandler.SearchTodos))       // TODO検索
		v1.GET("/todos/statistics", handler.ErrorHandler(todoHandler.GetStatistics)) // TODO統計情報

		// カテゴリーエンドポイント（Phase 2で追加）
		v1.POST("/categories", handler.ErrorHandler(categoryHandler.CreateCategory))      // カテゴリー作成
		v1.GET("/categories", handler.ErrorHandler(categoryHandler.GetCategories))        // カテゴリー一覧
		v1.GET("/categories/:id", handler.ErrorHandler(categoryHandler.GetCategory))      // カテゴリー取得
		v1.PUT("/categories/:id", handler.ErrorHandler(categoryHandler.UpdateCategory))   // カテゴリー更新
		v1.DELETE("/categories/:id", handler.ErrorHandler(categoryHandler.DeleteCategory)) // カテゴリー削除

		// 通知エンドポイント（Phase 4で追加）
		v1.GET("/notifications", handler.ErrorHandler(notificationHandler.GetNotifications))           // 通知一覧
		v1.GET("/notifications/unread", handler.ErrorHandler(notificationHandler.GetUnreadNotifications)) // 未読通知
		v1.PUT("/notifications/:id/read", handler.ErrorHandler(notificationHandler.MarkNotificationAsRead)) // 既読にする
		v1.PUT("/notifications/read-all", handler.ErrorHandler(notificationHandler.MarkAllNotificationsAsRead)) // 全て既読
		v1.DELETE("/notifications/:id", handler.ErrorHandler(notificationHandler.DeleteNotification))  // 通知削除

		// 管理者専用エンドポイント
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.AdminMiddleware())
		{
			adminRoutes.GET("/users", handler.ErrorHandler(adminHandler.GetAllUsers))
		}
	}

	return router
}
