package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
)

// NotificationHandler は通知関連のHTTPリクエストを処理
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler はNotificationHandlerの新しいインスタンスを作成
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications はユーザーの全通知を取得
func (h *NotificationHandler) GetNotifications(c *gin.Context) error {
	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	notifications, err := h.notificationService.GetNotifications(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, notifications)
	return nil
}

// GetUnreadNotifications は未読通知を取得
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) error {
	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	notifications, err := h.notificationService.GetUnreadNotifications(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, notifications)
	return nil
}

// MarkNotificationAsRead は通知を既読にする
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) error {
	notificationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	err = h.notificationService.MarkNotificationAsRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
	return nil
}

// MarkAllNotificationsAsRead は全ての通知を既読にする
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) error {
	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	err := h.notificationService.MarkAllNotificationsAsRead(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
	return nil
}

// DeleteNotification は通知を削除
func (h *NotificationHandler) DeleteNotification(c *gin.Context) error {
	notificationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	err = h.notificationService.DeleteNotification(c.Request.Context(), notificationID, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
	return nil
}

// StreamNotifications はSSEで未読通知をストリーミング配信
func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
	// SSEヘッダーの設定
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	ticker := time.NewTicker(5 * time.Second) // 5秒ごとにチェック
	defer ticker.Stop()

	// 初回送信
	notifications, err := h.notificationService.GetUnreadNotifications(c.Request.Context(), userID)
	if err == nil && len(notifications) > 0 {
		data, _ := json.Marshal(notifications)
		c.SSEvent("notification", string(data))
		c.Writer.Flush()
	}

	// 定期的に送信
	for {
		select {
		case <-ticker.C:
			notifications, err := h.notificationService.GetUnreadNotifications(c.Request.Context(), userID)
			if err != nil {
				continue
			}

			if len(notifications) > 0 {
				data, _ := json.Marshal(notifications)
				c.SSEvent("notification", string(data))
				c.Writer.Flush()
			}

		case <-c.Request.Context().Done():
			// クライアントが切断したら終了
			return
		}
	}
}
