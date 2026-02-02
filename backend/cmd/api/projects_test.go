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
	assert.Equal(t, 1, len(members))                   // オーナーのみ
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
