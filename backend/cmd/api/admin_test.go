package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// loginAndGetToken はログインしてトークンを取得するヘルパー関数
