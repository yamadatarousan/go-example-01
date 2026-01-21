package service

import (
	"context"

	"gin-quickstart/examples/backend/internal/domain"
	"gin-quickstart/examples/backend/internal/repository"
)

// NotificationService は通知関連のビジネスロジックを提供
type NotificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService はNotificationServiceの新しいインスタンスを作成
func NewNotificationService(notificationRepo repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification は通知を作成
func (s *NotificationService) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) (domain.Notification, error) {
	return s.notificationRepo.Create(ctx, input)
}

// GetNotifications はユーザーの全通知を取得
func (s *NotificationService) GetNotifications(ctx context.Context, userID int) ([]domain.Notification, error) {
	return s.notificationRepo.FindAll(ctx, userID)
}

// GetUnreadNotifications は未読通知を取得
func (s *NotificationService) GetUnreadNotifications(ctx context.Context, userID int) ([]domain.Notification, error) {
	return s.notificationRepo.FindUnread(ctx, userID)
}

// MarkNotificationAsRead は通知を既読にする
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notificationID, userID int) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationID, userID)
}

// MarkAllNotificationsAsRead は全ての通知を既読にする
func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, userID int) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification は通知を削除
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID, userID int) error {
	return s.notificationRepo.Delete(ctx, notificationID, userID)
}
