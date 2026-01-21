// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/backend/internal/service" → "gin-quickstart/backend/internal/service"

package handler

import (
	"net/http"
	"strconv"
	"time"

	"gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/service"

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
		return domain.ErrInvalidInput
	}

	var input struct {
		RemindAt string `json:"remind_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	// ISO8601形式の日時をパース
	remindAt, err := time.Parse(time.RFC3339, input.RemindAt)
	if err != nil {
		return domain.ErrInvalidInput
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
		return domain.ErrInvalidInput
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
		return domain.ErrInvalidInput
	}

	err = h.reminderService.DeleteReminder(c.Request.Context(), reminderID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted successfully"})
	return nil
}
