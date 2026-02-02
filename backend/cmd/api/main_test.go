package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gin-quickstart/backend/internal/config"
	"gin-quickstart/backend/internal/handler"
	"gin-quickstart/backend/internal/repository"
	"gin-quickstart/backend/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// getProjectRoot はgo.modファイルを探してプロジェクトルートを特定します。
func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	log.Fatal("Could not find project root (go.mod not found)")
	return ""
}

// getEnvOrDefault は環境変数を取得し、設定されていない場合はデフォルト値を返します。
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMainは、パッケージ内のテストが実行される前に一度だけ呼ばれる特別な関数です。
func TestMain(m *testing.M) {
	// CI環境かどうかを判定
	isCI := os.Getenv("CI") == "true"

	// プロジェクトルートを取得
	projectRoot := getProjectRoot()
	dockerComposePath := filepath.Join(projectRoot, "docker-compose.test.yml")
	migrationsPath := filepath.Join(projectRoot, "db", "migrations")
	seedDataPath := filepath.Join(projectRoot, "testdata", "seed.sql")

	// --- セットアップ ---
	if !isCI {
		// ローカル環境: docker-compose を起動
		log.Println("Spinning up test database...")
		// --waitフラグでhealthcheckが通るまで待機
		// ★ プロジェクトルート直下のdocker-compose.test.ymlを参照（常に同じファイルを使用）
		cmd := exec.Command("docker-compose", "-f", dockerComposePath, "up", "-d", "--wait")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Could not start test database: %v", err)
		}

		// deferでテスト終了時に必ずDBコンテナとボリュームを破棄する
		defer func() {
			log.Println("Tearing down test database and volumes...")
			cmd := exec.Command("docker-compose", "-f", dockerComposePath, "down", "-v")
			if err := cmd.Run(); err != nil {
				log.Printf("Could not stop test database: %v", err)
			}
		}()
	} else {
		log.Println("CI environment detected, skipping docker-compose setup...")
	}

	// DB接続情報を環境変数またはデフォルト値から取得
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5434")
	dbUser := getEnvOrDefault("DB_USER", "user")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "password")
	dbName := getEnvOrDefault("DB_NAME", "todo_test_db")

	// テスト用DBへの接続
	dsnForGo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	dsnForMigrate := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	// DBが完全に準備が整うまでリトライ
	for i := 0; i < 5; i++ {
		testDB, err = sql.Open("pgx", dsnForGo)
		if err == nil {
			if err = testDB.Ping(); err == nil {
				break
			}
		}
		log.Printf("Could not connect to test DB, retrying... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to test database after retries: %v", err)
	}

	if !isCI {
		// ローカル環境: マイグレーションとシードデータのロードを実行
		// マイグレーションの実行
		log.Println("Running migrations on test database...")
		// まず、既存のマイグレーションをすべてダウンさせ、スキーマをクリーンな状態に戻す
		// ★ プロジェクトルート直下のdb/migrationsを参照（常に同じディレクトリを使用）
		migrateDownCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", migrationsPath, "down", "-all")
		if output, err := migrateDownCmd.CombinedOutput(); err != nil {
			// エラーが発生しても続行（初回実行時など、ダウンするマイグレーションがない場合があるため）
			log.Printf("Could not run migrate down (may be normal on first run): %v\nOutput: %s", err, string(output))
		}

		// その後、すべてのマイグレーションをアップする
		migrateUpCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", migrationsPath, "up")
		if output, err := migrateUpCmd.CombinedOutput(); err != nil {
			log.Fatalf("Could not run migrations: %v\nOutput: %s", err, string(output))
		}

		// シードデータのロード
		log.Println("Loading seed data...")
		if err := loadSeedData(testDB, seedDataPath); err != nil {
			log.Fatalf("Could not load seed data: %v", err)
		}
	} else {
		log.Println("CI environment: migrations and seed data already loaded by GitHub Actions")
	}

	// --- テストの実行 ---
	code := m.Run()

	// --- 終了処理 ---
	os.Exit(code)
}

// setupTestRouterはテスト用のルーターを構築
func setupTestRouter(dbConn *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// テスト用の設定
	cfg := &config.Config{
		Database: config.DatabaseConfig{},
		Server: config.ServerConfig{
			Port:         "8080",
			AllowOrigins: []string{"*"},
		},
		JWT: config.JWTConfig{
			Secret: []byte("test-secret-key"),
		},
	}

	// 依存関係の構築
	todoRepo := repository.NewTodoRepository(dbConn)
	userRepo := repository.NewUserRepository(dbConn)
	categoryRepo := repository.NewCategoryRepository(dbConn)
	notificationRepo := repository.NewNotificationRepository(dbConn)
	reminderRepo := repository.NewReminderRepository(dbConn)
	projectRepo := repository.NewProjectRepository(dbConn)
	commentRepo := repository.NewCommentRepository(dbConn)
	assignmentRepo := repository.NewTodoAssignmentRepository(dbConn)
	refreshTokenRepo := repository.NewRefreshTokenRepository(dbConn)

	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWT.Secret)
	todoService := service.NewTodoService(todoRepo)
	adminService := service.NewAdminService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	reminderService := service.NewReminderService(reminderRepo, notificationRepo, todoRepo)
	projectService := service.NewProjectService(projectRepo)
	commentService := service.NewCommentService(commentRepo, todoRepo, projectRepo)
	assignmentService := service.NewTodoAssignmentService(assignmentRepo, todoRepo, projectRepo)

	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	todoOpenAPIHandler := handler.NewTodoOpenAPIAdapter(todoHandler)
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	reminderHandler := handler.NewReminderHandler(reminderService)
	projectHandler := handler.NewProjectHandler(projectService)
	commentHandler := handler.NewCommentHandler(commentService)
	assignmentHandler := handler.NewTodoAssignmentHandler(assignmentService)
	healthHandler := handler.NewHealthHandler(dbConn)

	// setupRouter関数を再利用
	return setupRouter(cfg, authService, userHandler, todoOpenAPIHandler, adminHandler, categoryHandler, notificationHandler, reminderHandler, projectHandler, commentHandler, assignmentHandler, healthHandler)
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
func loadSeedData(db *sql.DB, seedDataPath string) error {
	// ★ プロジェクトルート直下のtestdata/seed.sqlを参照（常に同じファイルを使用）
	seedSQL, err := os.ReadFile(seedDataPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(seedSQL))
	if err != nil {
		return err
	}
	return nil
}

// TestSignup はユーザー登録のテスト

// loginAndGetToken はログインしてトークンを取得するヘルパー関数
func loginAndGetToken(t *testing.T, router *gin.Engine) string {
	return loginAndGetTokenWithEmail(t, router, "user-test@example.com", "password123")
}

// loginAndGetTokenWithEmail は指定したメールアドレスでログインしてトークンを取得するヘルパー関数
func loginAndGetTokenWithEmail(t *testing.T, router *gin.Engine, email, password string) string {
	loginBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["access_token"] // Phase 6でaccess_tokenに変更
}

// createTodo はTODOを作成してIDを返すヘルパー関数

// createTodo はTODOを作成してIDを返すヘルパー関数
func createTodo(t *testing.T, router *gin.Engine, token, name string) int {
	todoBody := fmt.Sprintf(`{"name": "%s"}`, name)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return int(response["id"].(float64))
}

// completeTodo はTODOを完了するヘルパー関数
func completeTodo(t *testing.T, router *gin.Engine, token string, todoID int) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/complete", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// createCategory はカテゴリーを作成してIDを返すヘルパー関数
func createCategory(t *testing.T, router *gin.Engine, token, name, color string) int {
	categoryBody := fmt.Sprintf(`{"name": "%s", "color": "%s"}`, name, color)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBufferString(categoryBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return int(response["id"].(float64))
}

// createNotification は通知を作成してIDを返すヘルパー関数
func createNotification(t *testing.T, db *sql.DB, userID int, todoID *int, notifType, message string) int {
	query := `
		INSERT INTO notifications (user_id, todo_id, type, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int
	err := db.QueryRow(query, userID, todoID, notifType, message).Scan(&id)
	assert.NoError(t, err)
	return id
}
