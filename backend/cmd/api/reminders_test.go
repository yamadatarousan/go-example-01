package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gin-quickstart/backend/internal/repository"
	"gin-quickstart/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCreateReminder はリマインダー作成のテスト
func TestCreateReminder(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminder")

	// リマインダーを作成
	reminderBody := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var reminder map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminder)
	assert.NotNil(t, reminder["id"])
	assert.Equal(t, float64(todoID), reminder["todo_id"])
	assert.False(t, reminder["is_sent"].(bool))
}

// TestGetRemindersByTodoID はTODOのリマインダー一覧取得のテスト

func TestGetRemindersByTodoID(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminders list")

	// リマインダーを2つ作成
	reminderBody1 := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody1))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	reminderBody2 := `{"remind_at": "2027-01-15T15:00:00Z"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// リマインダー一覧を取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var reminders []map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminders)
	assert.GreaterOrEqual(t, len(reminders), 2)
}

// TestDeleteReminder はリマインダー削除のテスト

func TestDeleteReminder(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for reminder deletion")

	// リマインダーを作成
	reminderBody := `{"remind_at": "2026-12-31T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/todos/%d/reminders", todoID), bytes.NewBufferString(reminderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var reminder map[string]any
	json.Unmarshal(w.Body.Bytes(), &reminder)
	reminderID := int(reminder["id"].(float64))

	// リマインダーを削除
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/reminders/%d", reminderID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Reminder deleted successfully", response["message"])
}

// TestProcessPendingReminders はバックグラウンドワーカーの動作テスト

func TestProcessPendingReminders(t *testing.T) {
	router := setupTestRouter(testDB)
	token := loginAndGetToken(t, router)

	// テスト用のTODOを作成
	todoID := createTodo(t, router, token, "Test TODO for background worker")

	// 過去の時刻でリマインダーを作成（既に送信時刻を過ぎている）
	pastTime := time.Now().Add(-10 * time.Minute)
	query := `
		INSERT INTO reminders (todo_id, remind_at, is_sent)
		VALUES ($1, $2, FALSE)
		RETURNING id
	`
	var reminderID int
	err := testDB.QueryRow(query, todoID, pastTime).Scan(&reminderID)
	assert.NoError(t, err)

	// バックグラウンドワーカーのProcessPendingRemindersを手動実行
	reminderRepo := repository.NewReminderRepository(testDB)
	notificationRepo := repository.NewNotificationRepository(testDB)
	todoRepo := repository.NewTodoRepository(testDB)
	reminderService := service.NewReminderService(reminderRepo, notificationRepo, todoRepo)

	ctx := context.Background()
	err = reminderService.ProcessPendingReminders(ctx)
	assert.NoError(t, err)

	// 通知が作成されたことを確認
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var notifications []map[string]any
	json.Unmarshal(w.Body.Bytes(), &notifications)

	// リマインダーから作成された通知が含まれていることを確認
	found := false
	for _, n := range notifications {
		if n["type"] == "deadline_reminder" {
			found = true
			assert.Contains(t, n["message"], "Test TODO for background worker")
			break
		}
	}
	assert.True(t, found, "Reminder notification should be created")

	// リマインダーがis_sent=TRUEになっていることを確認
	var isSent bool
	query = `SELECT is_sent FROM reminders WHERE id = $1`
	err = testDB.QueryRow(query, reminderID).Scan(&isSent)
	assert.NoError(t, err)
	assert.True(t, isSent, "Reminder should be marked as sent")
}

// TestCreateProject はプロジェクト作成のテスト
