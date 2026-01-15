// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/service" → "gin-quickstart/internal/service"

package handler

import (
	"net/http"
	"strconv"
	"time"

	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/service"

	"github.com/gin-gonic/gin"
)

// ReminderHandler はリマインダー関連のHTTPリクエストを処理
type ReminderHandler struct {
	reminderService *service.ReminderService
}

// NewReminderHandler はReminderHandlerの新しいインスタンスを作成
func NewReminderHandler(reminderService *service.ReminderService) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

// CreateReminder はTODOにリマインダーを作成
func (h *ReminderHandler) CreateReminder(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return nil
	}

	var input struct {
		RemindAt string `json:"remind_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}

	// ISO8601形式の日時をパース
	remindAt, err := time.Parse(time.RFC3339, input.RemindAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO8601 format (e.g., 2026-01-15T15:00:00Z)"})
		return nil
	}

	reminderInput := domain.CreateReminderInput{
		TodoID:   todoID,
		RemindAt: remindAt,
	}

	reminder, err := h.reminderService.CreateReminder(c.Request.Context(), reminderInput)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, reminder)
	return nil
}

// GetRemindersByTodoID はTODOに紐づくリマインダー一覧を取得
func (h *ReminderHandler) GetRemindersByTodoID(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return nil
	}

	reminders, err := h.reminderService.GetRemindersByTodoID(c.Request.Context(), todoID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, reminders)
	return nil
}

// DeleteReminder はリマインダーを削除
func (h *ReminderHandler) DeleteReminder(c *gin.Context) error {
	reminderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
		return nil
	}

	err = h.reminderService.DeleteReminder(c.Request.Context(), reminderID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted successfully"})
	return nil
}
