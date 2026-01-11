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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// TestMainは、パッケージ内のテストが実行される前に一度だけ呼ばれる特別な関数です。
func TestMain(m *testing.M) {
	// --- セットアップ ---
	log.Println("Spinning up test database...")
	// --waitフラグでhealthcheckが通るまで待機
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yml", "up", "-d", "--wait")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not start test database: %v", err)
	}

	// deferでテスト終了時に必ずDBコンテナとボリュームを破棄する
	defer func() {
		log.Println("Tearing down test database and volumes...")
		cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yml", "down", "-v") // -v を追加
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
	migrateDownCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "db/migrations", "down", "-all") // -all に変更
	migrateDownCmd.Dir = ".."
	if output, err := migrateDownCmd.CombinedOutput(); err != nil {
		// エラーが発生しても続行（初回実行時など、ダウンするマイグレーションがない場合があるため）
		log.Printf("Could not run migrate down (may be normal on first run): %v\nOutput: %s", err, string(output))
	}

	// その後、すべてのマイグレーションをアップする
	migrateUpCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "db/migrations", "up")
	migrateUpCmd.Dir = ".."
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

// setupTestRouterはテスト用のDB接続を受け取るように変更
func setupTestRouter(dbConn *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// mainのdbではなく、引数で渡されたテスト用DB接続を使う
	repo := NewTodoRepository(dbConn)
	todoHandler := NewTodoHandler(repo)
	authHandler := NewAuthHandler(repo)
	adminHandler := NewAdminHandler(repo)

	router := gin.New()
	router.Use(cors.Default())

	router.POST("/signup", errorHandler(authHandler.signup))
	router.POST("/login", errorHandler(authHandler.login))

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware())
	{
		v1.GET("/todos", errorHandler(todoHandler.getTodos))
		v1.POST("/todos", errorHandler(todoHandler.createTodo))
		v1.PUT("/todos/:id", errorHandler(todoHandler.updateTodo))
		v1.DELETE("/todos/:id", errorHandler(todoHandler.deleteTodo))

		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(adminMiddleware())
		{
			adminRoutes.GET("/users", errorHandler(adminHandler.getAllUsers))
		}
	}
	return router
}

// TestUserFlowは、TestMainで準備されたテスト用DBを使って実行される
func TestUserFlow(t *testing.T) {
	// グローバルなtestDB接続を使ってルーターをセットアップ
	router := setupTestRouter(testDB)

	// --- 1. ログイン ---
	// TestMainでseed.sqlがロードされているため、既存のユーザーを使用
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]
	assert.NotEmpty(t, token)

	// --- 3. TODO作成 ---
	todoBody := `{"name": "Isolated Test Todo"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestUpdateTodoはTODO更新のテストです
func TestUpdateTodo(t *testing.T) {
	router := setupTestRouter(testDB)

	// ログイン
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]

	// TODO作成
	createBody := `{"name": "Original Todo"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	todoID := int(createResponse["id"].(float64))

	// TODO更新
	updateBody := `{"name": "Updated Todo"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/todos/%d", todoID), bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updateResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &updateResponse)
	assert.Equal(t, "Updated Todo", updateResponse["name"])
}

// TestDeleteTodoはTODO削除のテストです
func TestDeleteTodo(t *testing.T) {
	router := setupTestRouter(testDB)

	// ログイン
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]

	// TODO作成
	createBody := `{"name": "Todo to Delete"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	todoID := int(createResponse["id"].(float64))

	// TODO削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/todos/%d", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 削除後に再度GETして404が返ることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	// 削除したTODOが含まれていないことを確認
	for _, todo := range todos {
		assert.NotEqual(t, float64(todoID), todo["id"])
	}
}

// TestUpdateTodoNotFoundは存在しないTODOの更新テストです
func TestUpdateTodoNotFound(t *testing.T) {
	router := setupTestRouter(testDB)

	// ログイン
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]

	// 存在しないTODO IDで更新を試みる
	updateBody := `{"name": "Updated Todo"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/todos/99999", bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteTodoNotFoundは存在しないTODOの削除テストです
func TestDeleteTodoNotFound(t *testing.T) {
	router := setupTestRouter(testDB)

	// ログイン
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]

	// 存在しないTODO IDで削除を試みる
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/todos/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
func loadSeedData(db *sql.DB) error {
	seedSQL, err := os.ReadFile("../../go/testdata/seed.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(seedSQL))
	if err != nil {
		return err
	}
	return nil
}
