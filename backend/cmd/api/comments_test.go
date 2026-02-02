package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
