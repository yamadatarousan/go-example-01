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

	"gin-quickstart/examples/backend/internal/config"
	"gin-quickstart/examples/backend/internal/handler"
	"gin-quickstart/examples/backend/internal/middleware"
	"gin-quickstart/examples/backend/internal/repository"
	"gin-quickstart/examples/backend/internal/service"

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
	categoryRepo := repository.NewCategoryRepository(db)         // Phase 2で追加
	notificationRepo := repository.NewNotificationRepository(db) // Phase 4で追加
	reminderRepo := repository.NewReminderRepository(db)         // Phase 4で追加
	// tagRepo := repository.NewTagRepository(db)         // Phase 2で追加（Phase 3で使用）
	projectRepo := repository.NewProjectRepository(db)           // Phase 5で追加
	commentRepo := repository.NewCommentRepository(db)           // Phase 5で追加
	assignmentRepo := repository.NewTodoAssignmentRepository(db) // Phase 5で追加
	refreshTokenRepo := repository.NewRefreshTokenRepository(db) // Phase 6で追加

	// Service層
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWT.Secret)
	todoService := service.NewTodoService(todoRepo)
	adminService := service.NewAdminService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)                            // Phase 2で追加
	notificationService := service.NewNotificationService(notificationRepo)                           // Phase 4で追加
	reminderService := service.NewReminderService(reminderRepo, notificationRepo, todoRepo)               // Phase 4で追加
	projectService := service.NewProjectService(projectRepo)                                              // Phase 5で追加
	commentService := service.NewCommentService(commentRepo, todoRepo, projectRepo)                       // Phase 5で追加
	assignmentService := service.NewTodoAssignmentService(assignmentRepo, todoRepo, projectRepo)          // Phase 5で追加

	// Handler層
	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)             // Phase 2で追加
	notificationHandler := handler.NewNotificationHandler(notificationService) // Phase 4で追加
	reminderHandler := handler.NewReminderHandler(reminderService)             // Phase 4で追加
	projectHandler := handler.NewProjectHandler(projectService)                // Phase 5で追加
	commentHandler := handler.NewCommentHandler(commentService)                // Phase 5で追加
	assignmentHandler := handler.NewTodoAssignmentHandler(assignmentService)   // Phase 5で追加
	healthHandler := handler.NewHealthHandler(db)                              // Phase 6で追加

	// ルーターの設定
	router := setupRouter(cfg, authService, userHandler, todoHandler, adminHandler, categoryHandler, notificationHandler, reminderHandler, projectHandler, commentHandler, assignmentHandler, healthHandler)

	// バックグラウンドワーカーのコンテキスト
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// リマインダーワーカーを起動
	go startReminderWorker(workerCtx, reminderService)

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

	// バックグラウンドワーカーを停止
	workerCancel()

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
	categoryHandler *handler.CategoryHandler,         // Phase 2で追加
	notificationHandler *handler.NotificationHandler, // Phase 4で追加
	reminderHandler *handler.ReminderHandler,         // Phase 4で追加
	projectHandler *handler.ProjectHandler,           // Phase 5で追加
	commentHandler *handler.CommentHandler,           // Phase 5で追加
	assignmentHandler *handler.TodoAssignmentHandler, // Phase 5で追加
	healthHandler *handler.HealthHandler,             // Phase 6で追加
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
	router.Use(gin.Recovery())                               // panicからの回復
	router.Use(middleware.RequestIDMiddleware())             // リクエストIDの生成
	router.Use(gin.LoggerWithFormatter(logFormatter))        // ログ出力
	router.Use(middleware.RateLimiterMiddleware(100))        // Phase 6で追加: レート制限（100req/min）
	router.Use(middleware.SecurityHeadersMiddleware())       // Phase 6で追加: セキュリティヘッダー

	// ヘルスチェック
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/health", healthHandler.HealthCheck) // Phase 6で追加: 詳細なヘルスチェック

	// 認証エンドポイント
	router.POST("/signup", handler.ErrorHandler(userHandler.Signup))
	router.POST("/login", handler.ErrorHandler(userHandler.Login))
	router.POST("/api/v1/auth/refresh", handler.ErrorHandler(userHandler.RefreshToken))         // Phase 6で追加
	router.POST("/api/v1/auth/revoke", handler.ErrorHandler(userHandler.RevokeRefreshToken))    // Phase 6で追加

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
		v1.GET("/notifications/stream", notificationHandler.StreamNotifications)                       // SSEストリーム
		v1.PUT("/notifications/:id/read", handler.ErrorHandler(notificationHandler.MarkNotificationAsRead)) // 既読にする
		v1.PUT("/notifications/read-all", handler.ErrorHandler(notificationHandler.MarkAllNotificationsAsRead)) // 全て既読
		v1.DELETE("/notifications/:id", handler.ErrorHandler(notificationHandler.DeleteNotification))  // 通知削除

		// リマインダーエンドポイント（Phase 4で追加）
		v1.POST("/todos/:id/reminders", handler.ErrorHandler(reminderHandler.CreateReminder))        // リマインダー作成
		v1.GET("/todos/:id/reminders", handler.ErrorHandler(reminderHandler.GetRemindersByTodoID))   // リマインダー一覧
		v1.DELETE("/reminders/:id", handler.ErrorHandler(reminderHandler.DeleteReminder))            // リマインダー削除

		// プロジェクトエンドポイント（Phase 5で追加）
		v1.POST("/projects", handler.ErrorHandler(projectHandler.CreateProject))                       // プロジェクト作成
		v1.GET("/projects", handler.ErrorHandler(projectHandler.GetProjects))                          // プロジェクト一覧
		v1.GET("/projects/:id", handler.ErrorHandler(projectHandler.GetProject))                       // プロジェクト取得
		v1.PUT("/projects/:id", handler.ErrorHandler(projectHandler.UpdateProject))                    // プロジェクト更新
		v1.DELETE("/projects/:id", handler.ErrorHandler(projectHandler.DeleteProject))                 // プロジェクト削除
		v1.POST("/projects/:id/members", handler.ErrorHandler(projectHandler.AddMember))               // メンバー追加
		v1.GET("/projects/:id/members", handler.ErrorHandler(projectHandler.GetMembers))               // メンバー一覧
		v1.DELETE("/projects/:id/members/:userId", handler.ErrorHandler(projectHandler.RemoveMember))  // メンバー削除
		v1.PUT("/projects/:id/members/:userId/role", handler.ErrorHandler(projectHandler.UpdateMemberRole)) // 役割更新

		// TODO担当者エンドポイント（Phase 5で追加）
		v1.POST("/todos/:id/assignments", handler.ErrorHandler(assignmentHandler.AssignUser))         // 担当者割り当て
		v1.GET("/todos/:id/assignments", handler.ErrorHandler(assignmentHandler.GetAssignments))      // 担当者一覧
		v1.DELETE("/todos/:id/assignments/:userId", handler.ErrorHandler(assignmentHandler.UnassignUser)) // 担当者解除

		// コメントエンドポイント（Phase 5で追加）
		v1.POST("/todos/:id/comments", handler.ErrorHandler(commentHandler.CreateComment))            // コメント作成
		v1.GET("/todos/:id/comments", handler.ErrorHandler(commentHandler.GetCommentsByTodoID))       // コメント一覧
		v1.GET("/comments/:commentId", handler.ErrorHandler(commentHandler.GetComment))               // コメント取得
		v1.PUT("/comments/:commentId", handler.ErrorHandler(commentHandler.UpdateComment))            // コメント更新
		v1.DELETE("/comments/:commentId", handler.ErrorHandler(commentHandler.DeleteComment))         // コメント削除

		// 管理者専用エンドポイント
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.AdminMiddleware())
		{
			adminRoutes.GET("/users", handler.ErrorHandler(adminHandler.GetAllUsers))
		}
	}

	return router
}

// startReminderWorker はリマインダーをチェックして通知を作成するバックグラウンドワーカー
func startReminderWorker(ctx context.Context, reminderService *service.ReminderService) {
	ticker := time.NewTicker(1 * time.Minute) // 1分ごとに実行
	defer ticker.Stop()

	log.Println("リマインダーワーカーを起動しました")

	for {
		select {
		case <-ticker.C:
			// 送信待ちリマインダーを処理
			if err := reminderService.ProcessPendingReminders(ctx); err != nil {
				log.Printf("リマインダー処理エラー: %v", err)
			}
		case <-ctx.Done():
			// Graceful Shutdown時に停止
			log.Println("リマインダーワーカーを停止しました")
			return
		}
	}
}
