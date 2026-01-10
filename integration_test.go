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
	cmd := exec.Command("docker-compose", "-f", "./docker-compose.test.yml", "up", "-d", "--wait")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not start test database: %v", err)
	}

	// deferでテスト終了時に必ずDBコンテナとボリュームを破棄する
	defer func() {
		log.Println("Tearing down test database and volumes...")
		cmd := exec.Command("docker-compose", "-f", "./docker-compose.test.yml", "down", "-v") // -v を追加
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
	migrateDownCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "./db/migrations", "down", "-all") // -all に変更
	if output, err := migrateDownCmd.CombinedOutput(); err != nil {
		// エラーが発生しても続行（初回実行時など、ダウンするマイグレーションがない場合があるため）
		log.Printf("Could not run migrate down (may be normal on first run): %v\nOutput: %s", err, string(output))
	}

	// その後、すべてのマイグレーションをアップする
	migrateUpCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "./db/migrations", "up")
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

		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(adminMiddleware())
		{
			adminRoutes.GET("/users", errorHandler(adminHandler.getAllUsers))
		}
	}
	return router
}

// TestSignupは新規ユーザー登録のテストです
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
	assert.NotEmpty(t, response["id"])
	assert.Equal(t, "newuser@example.com", response["email"])
}

// TestLoginは既存ユーザーのログインテストです
func TestLogin(t *testing.T) {
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

// TestLoginInvalidCredentialsは誤った認証情報でのログインテストです
func TestLoginInvalidCredentials(t *testing.T) {
	router := setupTestRouter(testDB)

	loginBody := `{"email": "user-test@example.com", "password": "wrongpassword"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateTodoはTODO作成のテストです
func TestCreateTodo(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAsUser(t, router, "user-test@example.com", "password123")

	todoBody := `{"name": "New Test Todo"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["id"])
	assert.Equal(t, "New Test Todo", response["name"])
}

// TestGetTodosはTODO一覧取得のテストです
func TestGetTodos(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAsUser(t, router, "user-test@example.com", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var todos []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &todos)
	// シードデータで1つのTODOが作成されているはず
	assert.GreaterOrEqual(t, len(todos), 1)
}

// TestGetTodosUnauthorizedは認証なしでのTODO一覧取得テストです
func TestGetTodosUnauthorized(t *testing.T) {
	router := setupTestRouter(testDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAdminGetAllUsersは管理者による全ユーザー取得のテストです
func TestAdminGetAllUsers(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAsUser(t, router, "admin-test@example.com", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &users)
	// シードデータで2人のユーザーが作成されているはず
	assert.GreaterOrEqual(t, len(users), 2)
}

// TestAdminGetAllUsersForbiddenは一般ユーザーが管理者エンドポイントにアクセスするテストです
func TestAdminGetAllUsersForbidden(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAsUser(t, router, "user-test@example.com", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// loginAsUserはテスト用のヘルパー関数で、指定されたユーザーでログインしてトークンを返します
func loginAsUser(t *testing.T, router *gin.Engine, email, password string) string {
	loginBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	token := response["token"]
	assert.NotEmpty(t, token)
	return token
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
func loadSeedData(db *sql.DB) error {
	seedSQL, err := os.ReadFile("./testdata/seed.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(seedSQL))
	if err != nil {
		return err
	}
	return nil
}
