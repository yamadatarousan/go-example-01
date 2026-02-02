package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// TestGetTodos はTODO一覧取得のテスト
