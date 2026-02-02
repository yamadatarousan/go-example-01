package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetNotifications は通知一覧取得のテスト
func TestGetNotifications(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "Test notification message")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.GreaterOrEqual(t, len(notifications), 1)
}

// TestGetUnreadNotifications は未読通知取得のテスト

func TestGetUnreadNotifications(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の未読通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "Unread notification message")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.GreaterOrEqual(t, len(notifications), 1)
	// 全て未読であることを確認
	for _, n := range notifications {
		assert.False(t, n["is_read"].(bool))
	}
}

// TestMarkNotificationAsRead は通知を既読にするテスト

func TestMarkNotificationAsRead(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	notificationID := createNotification(t, testDB, 2, nil, "test_notification", "Test notification to mark as read")

	// 通知を既読にする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/notifications/%d/read", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Notification marked as read", response["message"])

	// 通知一覧で既読になっていることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// 作成した通知が既読になっていることを確認
	found := false
	for _, n := range notifications {
		if int(n["id"].(float64)) == notificationID {
			assert.True(t, n["is_read"].(bool))
			found = true
			break
		}
	}
	assert.True(t, found, "Notification should be found in the list")
}

// TestMarkAllNotificationsAsRead は全通知を既読にするテスト

func TestMarkAllNotificationsAsRead(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の複数の通知を作成
	createNotification(t, testDB, 2, nil, "test_notification", "First notification")
	createNotification(t, testDB, 2, nil, "test_notification", "Second notification")

	// 全通知を既読にする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "All notifications marked as read", response["message"])

	// 未読通知が0件であることを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var unreadNotifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &unreadNotifications)
	assert.Equal(t, 0, len(unreadNotifications))
}

// TestDeleteNotification は通知削除のテスト

func TestDeleteNotification(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用の通知を作成
	notificationID := createNotification(t, testDB, 2, nil, "test_notification", "Notification to delete")

	// 通知を削除
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/notifications/%d", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Notification deleted successfully", response["message"])

	// 通知一覧に存在しないことを確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// 削除した通知が存在しないことを確認
	for _, n := range notifications {
		assert.NotEqual(t, notificationID, int(n["id"].(float64)))
	}
}

// TestMarkNotificationAsReadUnauthorized は他ユーザーの通知を既読にできないテスト

func TestMarkNotificationAsReadUnauthorized(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 別ユーザー(admin)の通知を作成
	notificationID := createNotification(t, testDB, 1, nil, "test_notification", "Admin notification")

	// 通常ユーザーが管理者の通知を既読にしようとする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/notifications/%d/read", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 404 Not Foundが返ることを確認（権限エラーではなくリソースが見つからない扱い）
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteNotificationUnauthorized は他ユーザーの通知を削除できないテスト

func TestDeleteNotificationUnauthorized(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// 別ユーザー(admin)の通知を作成
	notificationID := createNotification(t, testDB, 1, nil, "test_notification", "Admin notification to delete")

	// 通常ユーザーが管理者の通知を削除しようとする
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/notifications/%d", notificationID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 404 Not Foundが返ることを確認
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// createNotification は通知を作成してIDを返すヘルパー関数
