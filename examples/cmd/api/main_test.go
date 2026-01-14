// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/config" → "gin-quickstart/internal/config"

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

// TestMainは、パッケージ内のテストが実行される前に一度だけ呼ばれる特別な関数です。
func TestMain(m *testing.M) {
	// --- セットアップ ---
	log.Println("Spinning up test database...")
	// --waitフラグでhealthcheckが通るまで待機
	// ★ プロジェクトルート直下のdocker-compose.test.ymlを参照（examples配下でも同じ）
	cmd := exec.Command("docker-compose", "-f", "../../../docker-compose.test.yml", "up", "-d", "--wait")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not start test database: %v", err)
	}

	// deferでテスト終了時に必ずDBコンテナとボリュームを破棄する
	defer func() {
		log.Println("Tearing down test database and volumes...")
		cmd := exec.Command("docker-compose", "-f", "../../../docker-compose.test.yml", "down", "-v")
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
	// ★ プロジェクトルート直下のdb/migrationsを参照（examples配下でも同じ）
	migrateDownCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "../../../db/migrations", "down", "-all")
	if output, err := migrateDownCmd.CombinedOutput(); err != nil {
		// エラーが発生しても続行（初回実行時など、ダウンするマイグレーションがない場合があるため）
		log.Printf("Could not run migrate down (may be normal on first run): %v\nOutput: %s", err, string(output))
	}

	// その後、すべてのマイグレーションをアップする
	migrateUpCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "../../../db/migrations", "up")
	if output, err := migrateUpCmd.CombinedOutput(); err != nil {
		log.Fatalf("Could not run migrations: %v\nOutput: %s", err, string(output))
	}

	// シードデータのロード
	log.Println("Loading seed data...")
	if err := loadSeedData(testDB); err != nil {
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

	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	todoService := service.NewTodoService(todoRepo)
	adminService := service.NewAdminService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	userHandler := handler.NewUserHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)
	adminHandler := handler.NewAdminHandler(adminService)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// setupRouter関数を再利用
	return setupRouter(cfg, authService, userHandler, todoHandler, adminHandler, categoryHandler)
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
// ★ プロジェクトルート直下のtestdata/seed.sqlを参照（examples配下でも同じ）
func loadSeedData(db *sql.DB) error {
	seedSQL, err := os.ReadFile("../../../testdata/seed.sql")
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
	assert.NotEmpty(t, response["token"])
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
	return response["token"]
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
