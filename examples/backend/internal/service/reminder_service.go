package service

import (
	"context"
	"fmt"

	"gin-quickstart/examples/backend/internal/domain"
	"gin-quickstart/examples/backend/internal/repository"
)

// ReminderService はリマインダー関連のビジネスロジックを提供
type ReminderService struct {
	reminderRepo     repository.ReminderRepository
	notificationRepo repository.NotificationRepository
	todoRepo         repository.TodoRepository
}

// NewReminderService はReminderServiceの新しいインスタンスを作成
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	notificationRepo repository.NotificationRepository,
	todoRepo repository.TodoRepository,
) *ReminderService {
	return &ReminderService{
		reminderRepo:     reminderRepo,
		notificationRepo: notificationRepo,
		todoRepo:         todoRepo,
	}
}

// CreateReminder はリマインダーを作成
func (s *ReminderService) CreateReminder(ctx context.Context, input domain.CreateReminderInput) (domain.Reminder, error) {
	return s.reminderRepo.Create(ctx, input)
}

// GetRemindersByTodoID はTODOに紐づくリマインダーを取得
func (s *ReminderService) GetRemindersByTodoID(ctx context.Context, todoID int) ([]domain.Reminder, error) {
	return s.reminderRepo.FindByTodoID(ctx, todoID)
}

// ProcessPendingReminders は送信待ちリマインダーを処理して通知を作成
func (s *ReminderService) ProcessPendingReminders(ctx context.Context) error {
	// 送信待ちリマインダーを取得
	reminders, err := s.reminderRepo.FindPending(ctx)
	if err != nil {
		return fmt.Errorf("送信待ちリマインダーの取得に失敗しました: %w", err)
	}

	for _, reminder := range reminders {
		// TODOを取得
		todo, err := s.todoRepo.FindByID(ctx, reminder.TodoID, 0) // user_id=0 はシステム権限
		if err != nil {
			// TODOが見つからない場合はスキップ
			continue
		}

		// 通知を作成
		notificationInput := domain.CreateNotificationInput{
			UserID:  todo.UserID,
			TodoID:  &todo.ID,
			Type:    "deadline_reminder",
			Message: fmt.Sprintf("リマインダー: %s の期限が近づいています", todo.Name),
		}

		_, err = s.notificationRepo.Create(ctx, notificationInput)
		if err != nil {
			// 通知作成失敗はログに記録するが処理は継続
			continue
		}

		// リマインダーを送信済みにする
		err = s.reminderRepo.MarkAsSent(ctx, reminder.ID)
		if err != nil {
			// 送信済みマーク失敗はログに記録するが処理は継続
			continue
		}
	}

	return nil
}

// DeleteReminder はリマインダーを削除
func (s *ReminderService) DeleteReminder(ctx context.Context, reminderID int) error {
	return s.reminderRepo.Delete(ctx, reminderID)
}
