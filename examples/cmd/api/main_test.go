// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/config" → "gin-quickstart/internal/config"

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gin-quickstart/examples/internal/config"
	"gin-quickstart/examples/internal/handler"
	"gin-quickstart/examples/internal/repository"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// getProjectRoot はgo.modファイルを探してプロジェクトルートを特定します。
// これにより、examples/cmd/api でもcmd/api でも同じコードで動作します。
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

// TestMainは、パッケージ内のテストが実行される前に一度だけ呼ばれる特別な関数です。
func TestMain(m *testing.M) {
	// プロジェクトルートを取得
	// ★ これにより、examples/cmd/api でもcmd/api でも同じコードで動作します
	projectRoot := getProjectRoot()
	dockerComposePath := filepath.Join(projectRoot, "docker-compose.test.yml")
	migrationsPath := filepath.Join(projectRoot, "db", "migrations")
	seedDataPath := filepath.Join(projectRoot, "testdata", "seed.sql")

	// --- セットアップ ---
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

	// テスト用DBへの接続
	dsnForGo := "host=localhost user=user password=password dbname=todo_test_db port=5434 sslmode=disable"
	dsnForMigrate := "postgres://user:password@localhost:5434/todo_test_db?sslmode=disable"

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
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	reminderHandler := handler.NewReminderHandler(reminderService)
	projectHandler := handler.NewProjectHandler(projectService)
	commentHandler := handler.NewCommentHandler(commentService)
	assignmentHandler := handler.NewTodoAssignmentHandler(assignmentService)
	healthHandler := handler.NewHealthHandler(dbConn)

	// setupRouter関数を再利用
	return setupRouter(cfg, authService, userHandler, todoHandler, adminHandler, categoryHandler, notificationHandler, reminderHandler, projectHandler, commentHandler, assignmentHandler, healthHandler)
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
// ★ プロジェクトルート直下のtestdata/seed.sqlを参照（examples配下でも同じ）
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

// ========== 1. ユーザー認証テスト ==========

// TestSignup はユーザー登録のテスト
func TestSignup(t *testing.T) {
	router := setupTestRouter(testDB)

	signupBody := `{"email": "newuser@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBufferString(signupBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "newuser@example.com", response["email"])
	assert.NotNil(t, response["id"])
}

// TestLoginSuccess はログイン成功のテスト
func TestLoginSuccess(t *testing.T) {
	router := setupTestRouter(testDB)

	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["access_token"])  // Phase 6でaccess_tokenに変更
	assert.NotEmpty(t, response["refresh_token"]) // Phase 6でrefresh_tokenを追加
}

// TestLoginFailure はログイン失敗のテスト
func TestLoginFailure(t *testing.T) {
	router := setupTestRouter(testDB)

	loginBody := `{"email": "user-test@example.com", "password": "wrongpassword"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRefreshToken はリフレッシュトークンのテスト（Phase 6で追加）
func TestRefreshToken(t *testing.T) {
	router := setupTestRouter(testDB)

	// 1. ログインしてrefresh_tokenを取得
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	refreshToken := loginResponse["refresh_token"]
	assert.NotEmpty(t, refreshToken)

	// 2. refresh_tokenを使って新しいaccess_tokenを取得
	refreshBody := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var refreshResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &refreshResponse)
	assert.NotEmpty(t, refreshResponse["access_token"])
}

// TestRevokeRefreshToken はリフレッシュトークン無効化のテスト（Phase 6で追加）
func TestRevokeRefreshToken(t *testing.T) {
	router := setupTestRouter(testDB)

	// 1. ログインしてrefresh_tokenを取得
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	refreshToken := loginResponse["refresh_token"]
	assert.NotEmpty(t, refreshToken)

	// 2. refresh_tokenを無効化
	revokeBody := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/revoke", bytes.NewBufferString(revokeBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 3. 無効化されたrefresh_tokenで新しいaccess_tokenを取得しようとする（失敗するはず）
	refreshBody := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ========== 2. TODO基本CRUD操作テスト ==========

// TestGetTodos はTODO一覧取得のテスト
func TestGetTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	assert.GreaterOrEqual(t, len(todos), 0)
}

// TestCreateTodoWithExtendedFields は拡張フィールドを含むTODO作成のテスト（Phase 2）
func TestCreateTodoWithExtendedFields(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// Phase 2で追加された拡張フィールドを含むTODO作成
	todoBody := `{
		"name": "Test Todo with Extended Fields",
		"priority": "high",
		"status": "todo",
		"description": "This is a test todo with extended fields",
		"due_date": "2026-12-31T23:59:59Z"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Todo with Extended Fields", response["name"])
	assert.Equal(t, "high", response["priority"])
	assert.Equal(t, "todo", response["status"])
	assert.NotNil(t, response["description"])
	assert.NotNil(t, response["due_date"])
}

// TestGetTodo は個別TODO取得のテスト
func TestGetTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODO作成
	todoID := createTodo(t, router, token, "Test Todo for Get")

	// 個別TODO取得
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Todo for Get", response["name"])
	assert.Equal(t, float64(todoID), response["id"])
}

// TestGetTodoNotFound は存在しないTODO取得のテスト
func TestGetTodoNotFound(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUpdateTodo はTODO更新のテスト
func TestUpdateTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODO作成
	todoID := createTodo(t, router, token, "Original Todo")

	// TODO更新
	updateBody := `{"name": "Updated Todo", "priority": "low"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/todos/%d", todoID), bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Updated Todo", response["name"])
	assert.Equal(t, "low", response["priority"])
}

// TestDeleteTodo はTODO削除のテスト
func TestDeleteTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODO作成
	todoID := createTodo(t, router, token, "Todo to Delete")

	// TODO削除
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/todos/%d", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 削除後にGETして404が返ることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== 3. TODO拡張機能テスト（Phase 2） ==========

// TestCompleteTodo はTODO完了のテスト
func TestCompleteTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODO作成
	todoID := createTodo(t, router, token, "Todo to Complete")

	// TODO完了
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/complete", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "done", response["status"])
}

// TestReopenTodo はTODO再開のテスト
func TestReopenTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODO作成して完了
	todoID := createTodo(t, router, token, "Todo to Reopen")
	completeTodo(t, router, token, todoID)

	// TODO再開
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reopen", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "todo", response["status"])
}

// TestGetOverdueTodos は期限切れTODO一覧取得のテスト
func TestGetOverdueTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 期限切れのTODOを作成（過去の日付）
	overdueTodoBody := `{
		"name": "Overdue Todo",
		"due_date": "2020-01-01T00:00:00Z",
		"status": "todo"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(overdueTodoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 期限切れTODO一覧取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/todos/overdue", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	assert.GreaterOrEqual(t, len(todos), 1)
}

// TestGetTodayTodos は今日のTODO一覧取得のテスト
func TestGetTodayTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 今日が期限のTODOを作成
	today := time.Now().Format(time.RFC3339)
	todayTodoBody := fmt.Sprintf(`{
		"name": "Today Todo",
		"due_date": "%s",
		"status": "todo"
	}`, today)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todayTodoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 今日のTODO一覧取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/todos/today", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	assert.GreaterOrEqual(t, len(todos), 1)
}

// TestGetThisWeekTodos は今週のTODO一覧取得のテスト
func TestGetThisWeekTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 今週が期限のTODOを作成（3日後）
	thisWeek := time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339)
	weekTodoBody := fmt.Sprintf(`{
		"name": "This Week Todo",
		"due_date": "%s",
		"status": "todo"
	}`, thisWeek)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(weekTodoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 今週のTODO一覧取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/todos/week", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	assert.GreaterOrEqual(t, len(todos), 1)
}

// ========== 4. カテゴリーCRUD操作テスト（Phase 2） ==========

// TestCreateCategory はカテゴリー作成のテスト
func TestCreateCategory(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	categoryBody := `{"name": "Work", "color": "#FF5733"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBufferString(categoryBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Work", response["name"])
	assert.Equal(t, "#FF5733", response["color"])
	assert.NotNil(t, response["id"])
}

// TestGetCategories はカテゴリー一覧取得のテスト
func TestGetCategories(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// カテゴリー作成
	createCategory(t, router, token, "Personal", "#00FF00")

	// カテゴリー一覧取得
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var categories []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &categories)
	assert.GreaterOrEqual(t, len(categories), 1)
}

// TestGetCategory は個別カテゴリー取得のテスト
func TestGetCategory(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// カテゴリー作成
	categoryID := createCategory(t, router, token, "Home", "#0000FF")

	// 個別カテゴリー取得
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/categories/%d", categoryID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Home", response["name"])
	assert.Equal(t, float64(categoryID), response["id"])
}

// TestUpdateCategory はカテゴリー更新のテスト
func TestUpdateCategory(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// カテゴリー作成
	categoryID := createCategory(t, router, token, "Original Category", "#AAAAAA")

	// カテゴリー更新
	updateBody := `{"name": "Updated Category", "color": "#BBBBBB"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/categories/%d", categoryID), bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Updated Category", response["name"])
	assert.Equal(t, "#BBBBBB", response["color"])
}

// TestDeleteCategory はカテゴリー削除のテスト
func TestDeleteCategory(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// カテゴリー作成
	categoryID := createCategory(t, router, token, "Category to Delete", "#CCCCCC")

	// カテゴリー削除
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/categories/%d", categoryID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// 削除後にGETして404が返ることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/categories/%d", categoryID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== 5. 管理者機能テスト ==========

// TestGetAllUsersAsAdmin は管理者による全ユーザー取得のテスト
func TestGetAllUsersAsAdmin(t *testing.T) {
	router := setupTestRouter(testDB)
	adminToken := loginAndGetTokenWithEmail(t, router, "admin-test@example.com", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &users)
	assert.GreaterOrEqual(t, len(users), 2) // 少なくとも2人のユーザー（admin, user）
}

// TestGetAllUsersAsForbidden は一般ユーザーによる全ユーザー取得の拒否テスト
func TestGetAllUsersAsForbidden(t *testing.T) {
	router := setupTestRouter(testDB)
	userToken := loginAndGetToken(t, router) // 一般ユーザーでログイン

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== ヘルパー関数 ==========

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

// ============================================================================
// Phase 3のテスト
// ============================================================================

// TestSearchTodos は検索エンドポイントのテスト
func TestSearchTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを複数作成
	todos := []string{
		`{"name": "High priority task", "priority": "high", "status": "todo"}`,
		`{"name": "Medium priority task", "priority": "medium", "status": "in_progress"}`,
		`{"name": "Low priority task", "priority": "low", "status": "done"}`,
		`{"name": "Search test task", "description": "This is a searchable description", "priority": "high", "status": "todo"}`,
	}

	for _, todoBody := range todos {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// テスト1: 優先度フィルター
	t.Run("Filter by priority", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/todos/search?priority=high", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.NotNil(t, result["todos"])
		todos := result["todos"].([]interface{})
		assert.GreaterOrEqual(t, len(todos), 2) // 最低2件のhigh priorityタスク
	})

	// テスト2: ステータスフィルター
	t.Run("Filter by status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/todos/search?status=done", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		todos := result["todos"].([]interface{})
		assert.GreaterOrEqual(t, len(todos), 1) // 最低1件のdoneタスク
	})

	// テスト3: 全文検索
	t.Run("Full-text search", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/todos/search?search=searchable", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.NotNil(t, result["todos"])
	})

	// テスト4: ページネーション
	t.Run("Pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/todos/search?page=1&limit=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(1), result["page"])
		assert.Equal(t, float64(2), result["limit"])
		todos := result["todos"].([]interface{})
		assert.LessOrEqual(t, len(todos), 2) // 最大2件
	})

	// テスト5: ソート順
	t.Run("Sort order", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/todos/search?sort=priority&order=desc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.NotNil(t, result["todos"])
	})
}

// TestGetStatistics は統計情報エンドポイントのテスト
func TestGetStatistics(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成（様々なステータス・優先度）
	todos := []string{
		`{"name": "Todo 1", "priority": "high", "status": "todo"}`,
		`{"name": "Todo 2", "priority": "medium", "status": "in_progress"}`,
		`{"name": "Todo 3", "priority": "low", "status": "done"}`,
		`{"name": "Todo 4", "priority": "high", "status": "todo"}`,
		`{"name": "Todo 5", "priority": "medium", "status": "done"}`,
	}

	for _, todoBody := range todos {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// 統計情報を取得
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos/statistics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &stats)

	// 総件数の確認
	assert.NotNil(t, stats["total_count"])
	totalCount := int(stats["total_count"].(float64))
	assert.GreaterOrEqual(t, totalCount, 5)

	// ステータス別カウントの確認
	statusCounts := stats["status_counts"].(map[string]interface{})
	assert.NotNil(t, statusCounts["todo"])
	assert.NotNil(t, statusCounts["in_progress"])
	assert.NotNil(t, statusCounts["done"])
	assert.GreaterOrEqual(t, int(statusCounts["todo"].(float64)), 2)
	assert.GreaterOrEqual(t, int(statusCounts["in_progress"].(float64)), 1)
	assert.GreaterOrEqual(t, int(statusCounts["done"].(float64)), 2)

	// 優先度別カウントの確認
	priorityCounts := stats["priority_counts"].(map[string]interface{})
	assert.NotNil(t, priorityCounts["high"])
	assert.NotNil(t, priorityCounts["medium"])
	assert.NotNil(t, priorityCounts["low"])
	assert.GreaterOrEqual(t, int(priorityCounts["high"].(float64)), 2)
	assert.GreaterOrEqual(t, int(priorityCounts["medium"].(float64)), 2)
	assert.GreaterOrEqual(t, int(priorityCounts["low"].(float64)), 1)

	// 期限関連カウントの確認
	assert.NotNil(t, stats["overdue_count"])
	assert.NotNil(t, stats["due_today_count"])
	assert.NotNil(t, stats["due_this_week_count"])
}

// ========== 9. 通知機能テスト（Phase 4） ==========

// TestGetNotifications は通知一覧取得のテスト
func TestGetNotifications(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "Test notification message")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.GreaterOrEqual(t, len(notifications), 1)
}

// TestGetUnreadNotifications は未読通知取得のテスト
func TestGetUnreadNotifications(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の未読通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "Unread notification message")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.GreaterOrEqual(t, len(notifications), 1)
	// 全て未読であることを確認
	for _, n := range notifications {
		assert.False(t, n["is_read"].(bool))
	}
}

// TestMarkNotificationAsRead は通知を既読にするテスト
func TestMarkNotificationAsRead(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	notificationID := createNotification(t, testDB, 2, nil, "test_notification", "Test notification to mark as read")

	// 通知を既読にする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/notifications/%d/read", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Notification marked as read", response["message"])

	// 通知一覧で既読になっていることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// 作成した通知が既読になっていることを確認
	found := false
	for _, n := range notifications {
		if int(n["id"].(float64)) == notificationID {
			assert.True(t, n["is_read"].(bool))
			found = true
			break
		}
	}
	assert.True(t, found, "Notification should be found in the list")
}

// TestMarkAllNotificationsAsRead は全通知を既読にするテスト
func TestMarkAllNotificationsAsRead(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の複数の通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "First notification")
	createNotification(t, testDB, 2, nil, "test_notification", "Second notification")

	// 全通知を既読にする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "All notifications marked as read", response["message"])

	// 未読通知が0件であることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var unreadNotifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &unreadNotifications)
	assert.Equal(t, 0, len(unreadNotifications))
}

// TestDeleteNotification は通知削除のテスト
func TestDeleteNotification(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	notificationID := createNotification(t, testDB, 2, nil, "test_notification", "Notification to delete")

	// 通知を削除
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/notifications/%d", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Notification deleted successfully", response["message"])

	// 通知一覧に存在しないことを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// 削除した通知が存在しないことを確認
	for _, n := range notifications {
		assert.NotEqual(t, notificationID, int(n["id"].(float64)))
	}
}

// TestMarkNotificationAsReadUnauthorized は他ユーザーの通知を既読にできないテスト
func TestMarkNotificationAsReadUnauthorized(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 別ユーザー(admin)の通知を作成
	notificationID := createNotification(t, testDB, 1, nil, "test_notification", "Admin notification")

	// 通常ユーザーが管理者の通知を既読にしようとする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/notifications/%d/read", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 404 Not Foundが返ることを確認（権限エラーではなくリソースが見つからない扱い）
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteNotificationUnauthorized は他ユーザーの通知を削除できないテスト
func TestDeleteNotificationUnauthorized(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 別ユーザー(admin)の通知を作成
	notificationID := createNotification(t, testDB, 1, nil, "test_notification", "Admin notification to delete")

	// 通常ユーザーが管理者の通知を削除しようとする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/notifications/%d", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 404 Not Foundが返ることを確認
	assert.Equal(t, http.StatusNotFound, w.Code)
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

// ========== 10. リマインダー機能テスト（Phase 4） ==========

// TestCreateReminder はリマインダー作成のテスト
func TestCreateReminder(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminder")

	// リマインダーを作成
	reminderBody := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var reminder map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminder)
	assert.NotNil(t, reminder["id"])
	assert.Equal(t, float64(todoID), reminder["todo_id"])
	assert.False(t, reminder["is_sent"].(bool))
}

// TestGetRemindersByTodoID はTODOのリマインダー一覧取得のテスト
func TestGetRemindersByTodoID(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminders list")

	// リマインダーを2つ作成
	reminderBody1 := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody1))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	reminderBody2 := `{"remind_at": "2027-01-15T15:00:00Z"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// リマインダー一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var reminders []map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminders)
	assert.GreaterOrEqual(t, len(reminders), 2)
}

// TestDeleteReminder はリマインダー削除のテスト
func TestDeleteReminder(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminder deletion")

	// リマインダーを作成
	reminderBody := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var reminder map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminder)
	reminderID := int(reminder["id"].(float64))

	// リマインダーを削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/reminders/%d", reminderID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Reminder deleted successfully", response["message"])
}

// ========== 11. バックグラウンドワーカーテスト（Phase 4） ==========

// TestProcessPendingReminders はバックグラウンドワーカーの動作テスト
func TestProcessPendingReminders(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for background worker")

	// 過去の時刻でリマインダーを作成（既に送信時刻を過ぎている）
	pastTime := time.Now().Add(-10 * time.Minute)
	query := `
		INSERT INTO reminders (todo_id, remind_at, is_sent)
		VALUES ($1, $2, FALSE)
		RETURNING id
	`
	var reminderID int
	err := testDB.QueryRow(query, todoID, pastTime).Scan(&reminderID)
	assert.NoError(t, err)

	// バックグラウンドワーカーのProcessPendingRemindersを手動実行
	reminderRepo := repository.NewReminderRepository(testDB)
	notificationRepo := repository.NewNotificationRepository(testDB)
	todoRepo := repository.NewTodoRepository(testDB)
	reminderService := service.NewReminderService(reminderRepo, notificationRepo, todoRepo)

	ctx := context.Background()
	err = reminderService.ProcessPendingReminders(ctx)
	assert.NoError(t, err)

	// 通知が作成されたことを確認
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]any
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// リマインダーから作成された通知が含まれていることを確認
	found := false
	for _, n := range notifications {
		if n["type"] == "deadline_reminder" {
			found = true
			assert.Contains(t, n["message"], "Test TODO for background worker")
			break
		}
	}
	assert.True(t, found, "Reminder notification should be created")

	// リマインダーがis_sent=TRUEになっていることを確認
	var isSent bool
	query = `SELECT is_sent FROM reminders WHERE id = $1`
	err = testDB.QueryRow(query, reminderID).Scan(&isSent)
	assert.NoError(t, err)
	assert.True(t, isSent, "Reminder should be marked as sent")
}

// ========== Phase 5: 共有・コラボレーション機能 ==========

// TestCreateProject はプロジェクト作成のテスト
func TestCreateProject(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	body := `{"name": "New Project", "description": "Project description"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	assert.Equal(t, "New Project", project["name"])
	assert.Equal(t, "Project description", project["description"])
}

// TestGetProjects はプロジェクト一覧取得のテスト
func TestGetProjects(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Test Project", "description": "Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// プロジェクト一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var projects []map[string]any
	json.Unmarshal(w.Body.Bytes(), &projects)
	assert.Greater(t, len(projects), 0)
}

// TestGetProject は単一プロジェクト取得のテスト
func TestGetProject(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Get Project Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// プロジェクトを取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &project)
	assert.Equal(t, "Get Project Test", project["name"])
}

// TestUpdateProject はプロジェクト更新のテスト
func TestUpdateProject(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Old Name"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// プロジェクトを更新
	body = `{"name": "New Name", "description": "Updated"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%d", projectID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &project)
	assert.Equal(t, "New Name", project["name"])
}

// TestDeleteProject はプロジェクト削除のテスト
func TestDeleteProject(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Delete Me"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// プロジェクトを削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestAddMember はプロジェクトメンバー追加のテスト
func TestAddMember(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Team Project"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// メンバーを追加（user3を追加）
	body = `{"user_id": 3, "role": "member"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%d/members", projectID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestGetMembers はプロジェクトメンバー一覧取得のテスト
func TestGetMembers(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Members Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// メンバー一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%d/members", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var members []map[string]any
	json.Unmarshal(w.Body.Bytes(), &members)
	assert.Equal(t, 1, len(members)) // オーナーのみ
	assert.Equal(t, float64(2), members[0]["user_id"]) // user2がオーナー
	assert.Equal(t, "owner", members[0]["role"])
}

// TestRemoveMember はプロジェクトメンバー削除のテスト
func TestRemoveMember(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Remove Member Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// メンバーを追加
	body = `{"user_id": 3, "role": "member"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%d/members", projectID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// メンバーを削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%d/members/3", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestUpdateMemberRole はプロジェクトメンバー役割更新のテスト
func TestUpdateMemberRole(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// プロジェクトを作成
	body := `{"name": "Role Update Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var project map[string]any
	json.Unmarshal(w.Body.Bytes(), &project)
	projectID := int(project["id"].(float64))

	// メンバーを追加
	body = `{"user_id": 3, "role": "member"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%d/members", projectID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 役割を更新
	body = `{"role": "admin"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%d/members/3/role", projectID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAssignUser はTODO担当者割り当てのテスト
func TestAssignUser(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// TODOを取得（seed.sqlのTODO ID=1を使用）
	todoID := 1

	// 担当者を割り当て
	body := `{"user_id": 1}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/assignments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var assignment map[string]any
	json.Unmarshal(w.Body.Bytes(), &assignment)
	assert.Equal(t, float64(todoID), assignment["todo_id"])
	assert.Equal(t, float64(1), assignment["user_id"])
}

// TestGetAssignments はTODO担当者一覧取得のテスト
func TestGetAssignments(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// 担当者を割り当て
	body := `{"user_id": 1}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/assignments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 担当者一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d/assignments", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var assignments []map[string]any
	json.Unmarshal(w.Body.Bytes(), &assignments)
	assert.Greater(t, len(assignments), 0)
}

// TestUnassignUser はTODO担当者解除のテスト
func TestUnassignUser(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// 担当者を割り当て
	body := `{"user_id": 1}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/assignments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 担当者を解除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/todos/%d/assignments/1", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestCreateComment はコメント作成のテスト
func TestCreateComment(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// コメントを作成
	body := `{"content": "This is a test comment"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var comment map[string]any
	json.Unmarshal(w.Body.Bytes(), &comment)
	assert.Equal(t, "This is a test comment", comment["content"])
	assert.Equal(t, float64(todoID), comment["todo_id"])
}

// TestGetCommentsByTodoID はコメント一覧取得のテスト
func TestGetCommentsByTodoID(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// コメントを作成
	body := `{"content": "Comment 1"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// コメント一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var comments []map[string]any
	json.Unmarshal(w.Body.Bytes(), &comments)
	assert.Greater(t, len(comments), 0)
}

// TestGetComment は単一コメント取得のテスト
func TestGetComment(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// コメントを作成
	body := `{"content": "Get me"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var comment map[string]any
	json.Unmarshal(w.Body.Bytes(), &comment)
	commentID := int(comment["id"].(float64))

	// コメントを取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/comments/%d", commentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &comment)
	assert.Equal(t, "Get me", comment["content"])
}

// TestUpdateComment はコメント更新のテスト
func TestUpdateComment(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// コメントを作成
	body := `{"content": "Old content"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var comment map[string]any
	json.Unmarshal(w.Body.Bytes(), &comment)
	commentID := int(comment["id"].(float64))

	// コメントを更新
	body = `{"content": "New content"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/comments/%d", commentID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &comment)
	assert.Equal(t, "New content", comment["content"])
}

// TestDeleteComment はコメント削除のテスト
func TestDeleteComment(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoID := 1

	// コメントを作成
	body := `{"content": "Delete me"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/comments", todoID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var comment map[string]any
	json.Unmarshal(w.Body.Bytes(), &comment)
	commentID := int(comment["id"].(float64))

	// コメントを削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/comments/%d", commentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
