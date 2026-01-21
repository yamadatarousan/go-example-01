package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gin-quickstart/backend/internal/domain"
)

// notificationRepository はNotificationRepositoryインターフェースの実装
type notificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository はNotificationRepositoryの新しいインスタンスを作成
func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create は通知を作成
func (r *notificationRepository) Create(ctx context.Context, input domain.CreateNotificationInput) (domain.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, todo_id, type, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, todo_id, type, message, is_read, created_at
	`

	var notification domain.Notification
	err := r.db.QueryRowContext(ctx, query, input.UserID, input.TodoID, input.Type, input.Message).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.TodoID,
		&notification.Type,
		&notification.Message,
		&notification.IsRead,
		&notification.CreatedAt,
	)
	if err != nil {
		return notification, fmt.Errorf("通知の作成に失敗しました: %w", err)
	}

	return notification, nil
}

// FindAll はユーザーの全通知を取得
func (r *notificationRepository) FindAll(ctx context.Context, userID int) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, todo_id, type, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("通知一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.TodoID, &n.Type, &n.Message, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("通知のスキャンに失敗しました: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("通知一覧の取得に失敗しました: %w", err)
	}

	return notifications, nil
}

// FindUnread は未読通知を取得
func (r *notificationRepository) FindUnread(ctx context.Context, userID int) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, todo_id, type, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1 AND is_read = FALSE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("未読通知の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.TodoID, &n.Type, &n.Message, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("通知のスキャンに失敗しました: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("未読通知の取得に失敗しました: %w", err)
	}

	return notifications, nil
}

// MarkAsRead は通知を既読にする
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, userID int) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("既読の更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// MarkAllAsRead は全ての通知を既読にする
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID int) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("全既読の更新に失敗しました: %w", err)
	}

	return nil
}

// Delete は通知を削除
func (r *notificationRepository) Delete(ctx context.Context, notificationID, userID int) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("通知の削除に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
