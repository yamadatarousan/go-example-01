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

// TestGetAllUsersAsAdmin は管理者による全ユーザー取得のテスト
