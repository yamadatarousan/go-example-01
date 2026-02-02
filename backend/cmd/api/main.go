package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gin-quickstart/backend/internal/config"
	appdb "gin-quickstart/backend/internal/db"
	"gin-quickstart/backend/internal/handler"
	"gin-quickstart/backend/internal/middleware"
	openapi "gin-quickstart/backend/internal/openapi/gen"
	"gin-quickstart/backend/internal/repository"
	"gin-quickstart/backend/internal/service"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	oapigin "github.com/oapi-codegen/gin-middleware"
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

	// 開発環境のみ自動マイグレーション（AUTO_MIGRATE=true）
	if cfg.Database.AutoMigrate {
		if err := runMigrations(db, cfg.Database); err != nil {
			log.Fatalf("自動マイグレーションに失敗しました: %v", err)
		}
	}

	// 必須テーブルの存在チェック（起動時に不足を検知）
	if err := checkRequiredTables(db, appdb.RequiredTables); err != nil {
		log.Fatalf("必須テーブルが不足しています: %v", err)
	}

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
	categoryService := service.NewCategoryService(categoryRepo)                                  // Phase 2で追加
	notificationService := service.NewNotificationService(notificationRepo)                      // Phase 4で追加
	reminderService := service.NewReminderService(reminderRepo, notificationRepo, todoRepo)      // Phase 4で追加
	projectService := service.NewProjectService(projectRepo)                                     // Phase 5で追加
	commentService := service.NewCommentService(commentRepo, todoRepo, projectRepo)              // Phase 5で追加
	assignmentService := service.NewTodoAssignmentService(assignmentRepo, todoRepo, projectRepo) // Phase 5で追加

	// Handler層
	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	todoOpenAPIHandler := handler.NewTodoOpenAPIAdapter(todoHandler)
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)             // Phase 2で追加
	notificationHandler := handler.NewNotificationHandler(notificationService) // Phase 4で追加
	reminderHandler := handler.NewReminderHandler(reminderService)             // Phase 4で追加
	projectHandler := handler.NewProjectHandler(projectService)                // Phase 5で追加
	commentHandler := handler.NewCommentHandler(commentService)                // Phase 5で追加
	assignmentHandler := handler.NewTodoAssignmentHandler(assignmentService)   // Phase 5で追加
	healthHandler := handler.NewHealthHandler(db)                              // Phase 6で追加

	// ルーターの設定
	router := setupRouter(cfg, authService, userHandler, todoOpenAPIHandler, adminHandler, categoryHandler, notificationHandler, reminderHandler, projectHandler, commentHandler, assignmentHandler, healthHandler)

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

// runMigrations は起動時にマイグレーションを実行する
func runMigrations(db *sql.DB, dbConfig config.DatabaseConfig) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}
	migrationsPath := filepath.Join(projectRoot, "db", "migrations")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("マイグレーション用ドライバの初期化に失敗しました: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("マイグレーションの初期化に失敗しました: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}

	log.Println("マイグレーションを適用しました")
	return nil
}

// checkRequiredTables は必須テーブルが存在するか確認する
func checkRequiredTables(db *sql.DB, tables []string) error {
	missing := make([]string, 0)
	for _, table := range tables {
		var regclass sql.NullString
		err := db.QueryRow("SELECT to_regclass($1)", "public."+table).Scan(&regclass)
		if err != nil {
			return fmt.Errorf("テーブル存在チェックに失敗しました: %w", err)
		}
		if !regclass.Valid {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("不足テーブル: %s", strings.Join(missing, ", "))
	}
	return nil
}

// getProjectRoot は go.mod を探してプロジェクトルートを特定する
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("作業ディレクトリの取得に失敗しました: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("プロジェクトルートが見つかりません（go.modがありません）")
}

// setupRouter はGinルーターを設定
func setupRouter(
	cfg *config.Config,
	authService *service.AuthService,
	userHandler *handler.UserHandler,
	todoOpenAPIHandler *handler.TodoOpenAPIAdapter,
	adminHandler *handler.AdminHandler,
	categoryHandler *handler.CategoryHandler, // Phase 2で追加
	notificationHandler *handler.NotificationHandler, // Phase 4で追加
	reminderHandler *handler.ReminderHandler, // Phase 4で追加
	projectHandler *handler.ProjectHandler, // Phase 5で追加
	commentHandler *handler.CommentHandler, // Phase 5で追加
	assignmentHandler *handler.TodoAssignmentHandler, // Phase 5で追加
	healthHandler *handler.HealthHandler, // Phase 6で追加
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
	router.Use(gin.Recovery())                         // panicからの回復
	router.Use(middleware.RequestIDMiddleware())       // リクエストIDの生成
	router.Use(gin.LoggerWithFormatter(logFormatter))  // ログ出力
	router.Use(middleware.RateLimiterMiddleware(100))  // Phase 6で追加: レート制限（100req/min）
	router.Use(middleware.SecurityHeadersMiddleware()) // Phase 6で追加: セキュリティヘッダー

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
	router.POST("/api/v1/auth/refresh", handler.ErrorHandler(userHandler.RefreshToken))      // Phase 6で追加
	router.POST("/api/v1/auth/revoke", handler.ErrorHandler(userHandler.RevokeRefreshToken)) // Phase 6で追加

	// 認証が必要なエンドポイント
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(authService))
	{
		// カテゴリーエンドポイント（Phase 2で追加）
		v1.POST("/categories", handler.ErrorHandler(categoryHandler.CreateCategory))       // カテゴリー作成
		v1.GET("/categories", handler.ErrorHandler(categoryHandler.GetCategories))         // カテゴリー一覧
		v1.GET("/categories/:id", handler.ErrorHandler(categoryHandler.GetCategory))       // カテゴリー取得
		v1.PUT("/categories/:id", handler.ErrorHandler(categoryHandler.UpdateCategory))    // カテゴリー更新
		v1.DELETE("/categories/:id", handler.ErrorHandler(categoryHandler.DeleteCategory)) // カテゴリー削除

		// 通知エンドポイント（Phase 4で追加）
		v1.GET("/notifications", handler.ErrorHandler(notificationHandler.GetNotifications))                    // 通知一覧
		v1.GET("/notifications/unread", handler.ErrorHandler(notificationHandler.GetUnreadNotifications))       // 未読通知
		v1.GET("/notifications/stream", notificationHandler.StreamNotifications)                                // SSEストリーム
		v1.PUT("/notifications/:id/read", handler.ErrorHandler(notificationHandler.MarkNotificationAsRead))     // 既読にする
		v1.PUT("/notifications/read-all", handler.ErrorHandler(notificationHandler.MarkAllNotificationsAsRead)) // 全て既読
		v1.DELETE("/notifications/:id", handler.ErrorHandler(notificationHandler.DeleteNotification))           // 通知削除

		// リマインダーエンドポイント（Phase 4で追加）
		v1.POST("/todos/:id/reminders", handler.ErrorHandler(reminderHandler.CreateReminder))      // リマインダー作成
		v1.GET("/todos/:id/reminders", handler.ErrorHandler(reminderHandler.GetRemindersByTodoID)) // リマインダー一覧
		v1.DELETE("/reminders/:id", handler.ErrorHandler(reminderHandler.DeleteReminder))          // リマインダー削除

		// プロジェクトエンドポイント（Phase 5で追加）
		v1.POST("/projects", handler.ErrorHandler(projectHandler.CreateProject))                            // プロジェクト作成
		v1.GET("/projects", handler.ErrorHandler(projectHandler.GetProjects))                               // プロジェクト一覧
		v1.GET("/projects/:id", handler.ErrorHandler(projectHandler.GetProject))                            // プロジェクト取得
		v1.PUT("/projects/:id", handler.ErrorHandler(projectHandler.UpdateProject))                         // プロジェクト更新
		v1.DELETE("/projects/:id", handler.ErrorHandler(projectHandler.DeleteProject))                      // プロジェクト削除
		v1.POST("/projects/:id/members", handler.ErrorHandler(projectHandler.AddMember))                    // メンバー追加
		v1.GET("/projects/:id/members", handler.ErrorHandler(projectHandler.GetMembers))                    // メンバー一覧
		v1.DELETE("/projects/:id/members/:userId", handler.ErrorHandler(projectHandler.RemoveMember))       // メンバー削除
		v1.PUT("/projects/:id/members/:userId/role", handler.ErrorHandler(projectHandler.UpdateMemberRole)) // 役割更新

		// TODO担当者エンドポイント（Phase 5で追加）
		v1.POST("/todos/:id/assignments", handler.ErrorHandler(assignmentHandler.AssignUser))             // 担当者割り当て
		v1.GET("/todos/:id/assignments", handler.ErrorHandler(assignmentHandler.GetAssignments))          // 担当者一覧
		v1.DELETE("/todos/:id/assignments/:userId", handler.ErrorHandler(assignmentHandler.UnassignUser)) // 担当者解除

		// コメントエンドポイント（Phase 5で追加）
		v1.POST("/todos/:id/comments", handler.ErrorHandler(commentHandler.CreateComment))      // コメント作成
		v1.GET("/todos/:id/comments", handler.ErrorHandler(commentHandler.GetCommentsByTodoID)) // コメント一覧
		v1.GET("/comments/:commentId", handler.ErrorHandler(commentHandler.GetComment))         // コメント取得
		v1.PUT("/comments/:commentId", handler.ErrorHandler(commentHandler.UpdateComment))      // コメント更新
		v1.DELETE("/comments/:commentId", handler.ErrorHandler(commentHandler.DeleteComment))   // コメント削除

		// 管理者専用エンドポイント
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.AdminMiddleware())
		{
			adminRoutes.GET("/users", handler.ErrorHandler(adminHandler.GetAllUsers))
		}
	}

	openapiSpec, err := loadOpenAPISpec()
	if err != nil {
		log.Fatalf("OpenAPI仕様の読み込みに失敗しました: %v", err)
	}

	openapiSpec.Servers = nil
	openapiRouter, err := gorillamux.NewRouter(openapiSpec)
	if err != nil {
		log.Fatalf("OpenAPIルーターの初期化に失敗しました: %v", err)
	}
	requestValidator := oapigin.OapiRequestValidatorWithOptions(openapiSpec, &oapigin.Options{
		SilenceServersWarning: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			response := gin.H{
				"error":   http.StatusText(statusCode),
				"details": message,
			}
			if os.Getenv("OPENAPI_DEBUG") == "true" {
				response["error_source"] = "openapi"
			}
			c.JSON(statusCode, response)
		},
	})

	todoRoutes := router.Group("")
	todoRoutes.Use(requestValidator)
	todoRoutes.Use(responseValidationMiddleware(openapiRouter))
	openapi.RegisterHandlersWithOptions(todoRoutes, todoOpenAPIHandler, openapi.GinServerOptions{
		Middlewares: []openapi.MiddlewareFunc{
			func(c *gin.Context) {
				middleware.AuthMiddleware(authService)(c)
			},
		},
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{
				"error":   "Bad Request",
				"details": err.Error(),
			})
		},
	})

	return router
}

type responseBufferWriter struct {
	gin.ResponseWriter
	body        bytes.Buffer // レスポンスボディをバッファ
	statusCode  int          // ステータスコードを保持
	wroteHeader bool         // ヘッダ書き込み済みか
	size        int          // 書き込みサイズ
}

func (w *responseBufferWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
	w.wroteHeader = true
}

func (w *responseBufferWriter) WriteHeaderNow() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
}

func (w *responseBufferWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	w.body.Write(data)
	w.size += len(data)
	return len(data), nil
}

func (w *responseBufferWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *responseBufferWriter) Status() int {
	if w.wroteHeader {
		return w.statusCode
	}
	return http.StatusOK
}

func (w *responseBufferWriter) Size() int {
	return w.size
}

func (w *responseBufferWriter) Written() bool {
	return w.wroteHeader || w.size > 0
}

func (w *responseBufferWriter) Flush() {
	// no-op: buffered
}

func responseValidationMiddleware(router routers.Router) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 通常はレスポンスが書き込みと同時に送信されるため、
		// 検証に必要な「ボディ全体」を事前に取得できない。
		// そのため一旦バッファに溜めてから検証し、問題なければ返却する。
		originalWriter := c.Writer
		bufferedWriter := &responseBufferWriter{ResponseWriter: originalWriter}
		c.Writer = bufferedWriter

		c.Next()

		statusCode := bufferedWriter.Status()
		bodyBytes := bufferedWriter.body.Bytes()

		// 既にエラー応答済みの場合は検証せずに返却する
		if c.IsAborted() {
			c.Writer = originalWriter
			originalWriter.WriteHeader(statusCode)
			if len(bodyBytes) > 0 {
				_, _ = originalWriter.Write(bodyBytes)
			}
			return
		}

		// OpenAPIルートへマッチさせてレスポンス検証を行う
		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			c.Writer = originalWriter
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("レスポンス検証エラー: route not found: %v", err)
			return
		}

		validationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request:    c.Request,
				PathParams: pathParams,
				Route:      route,
			},
			Status: statusCode,
			Header: originalWriter.Header(),
			Body:   io.NopCloser(bytes.NewReader(bodyBytes)),
		}

		// 契約違反時は500で返し、ログに詳細を残す
		if err := openapi3filter.ValidateResponse(c.Request.Context(), validationInput); err != nil {
			requestID, _ := c.Get("RequestID")
			userID := ""
			if claimsVal, ok := c.Get("claims"); ok {
				if claims, ok := claimsVal.(*service.AppClaims); ok {
					userID = claims.Subject
				}
			}
			log.Printf("レスポンス検証エラー: request_id=%v method=%s path=%s status=%d user_id=%s error=%v",
				requestID, c.Request.Method, c.Request.URL.Path, statusCode, userID, err)

			c.Writer = originalWriter
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		// 検証OKなら元のレスポンスを返却
		c.Writer = originalWriter
		originalWriter.WriteHeader(statusCode)
		if len(bodyBytes) > 0 {
			_, _ = originalWriter.Write(bodyBytes)
		}
	}
}

func loadOpenAPISpec() (*openapi3.T, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	specPath := filepath.Join(projectRoot, "openapi", "openapi.yaml")
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	return loader.LoadFromFile(specPath)
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
