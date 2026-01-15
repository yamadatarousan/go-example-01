package handler

import (
	"net/http"
	"strconv"

	"gin-quickstart/internal/service"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return nil
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return nil
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
