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
