package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

// TestCreateTodoValidationMissingName は必須フィールド欠落の検証テスト
func TestCreateTodoValidationMissingName(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoBody := `{"priority": "high", "status": "todo"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateTodoValidationInvalidEnum はenum違反の検証テスト
func TestCreateTodoValidationInvalidEnum(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)
	t.Setenv("OPENAPI_DEBUG", "true")

	todoBody := `{"name": "Invalid Enum", "priority": "urgent", "status": "todo"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "openapi", response["error_source"])
}

// TestCreateTodoValidationTypeMismatch は型不一致の検証テスト
func TestCreateTodoValidationTypeMismatch(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)
	t.Setenv("OPENAPI_DEBUG", "true")

	todoBody := `{"name": "Type Mismatch", "tag_ids": "not-an-array"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "openapi", response["error_source"])
}

// TestCreateTodoValidationNullable はnullable項目の許容テスト
func TestCreateTodoValidationNullable(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	todoBody := `{"name": "Nullable Todo", "description": null, "status": "todo", "priority": "low"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
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

// TestCreateCategory はカテゴリー作成のテスト

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

// TestGetNotifications は通知一覧取得のテスト
