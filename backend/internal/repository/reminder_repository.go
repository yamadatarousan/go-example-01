package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gin-quickstart/backend/internal/domain"
)

// reminderRepository はReminderRepositoryインターフェースの実装
type reminderRepository struct {
	db *sql.DB
}

// NewReminderRepository はReminderRepositoryの新しいインスタンスを作成
func NewReminderRepository(db *sql.DB) ReminderRepository {
	return &reminderRepository{db: db}
}

// Create はリマインダーを作成
func (r *reminderRepository) Create(ctx context.Context, input domain.CreateReminderInput) (domain.Reminder, error) {
	query := `
		INSERT INTO reminders (todo_id, remind_at)
		VALUES ($1, $2)
		RETURNING id, todo_id, remind_at, is_sent, created_at
	`

	var reminder domain.Reminder
	err := r.db.QueryRowContext(ctx, query, input.TodoID, input.RemindAt).Scan(
		&reminder.ID,
		&reminder.TodoID,
		&reminder.RemindAt,
		&reminder.IsSent,
		&reminder.CreatedAt,
	)
	if err != nil {
		return reminder, fmt.Errorf("リマインダーの作成に失敗しました: %w", err)
	}

	return reminder, nil
}

// FindByTodoID はTODOに紐づくリマインダーを取得
func (r *reminderRepository) FindByTodoID(ctx context.Context, todoID int) ([]domain.Reminder, error) {
	query := `
		SELECT id, todo_id, remind_at, is_sent, created_at
		FROM reminders
		WHERE todo_id = $1
		ORDER BY remind_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, todoID)
	if err != nil {
		return nil, fmt.Errorf("リマインダー一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var reminders []domain.Reminder
	for rows.Next() {
		var reminder domain.Reminder
		err := rows.Scan(&reminder.ID, &reminder.TodoID, &reminder.RemindAt, &reminder.IsSent, &reminder.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("リマインダーのスキャンに失敗しました: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("リマインダー一覧の取得に失敗しました: %w", err)
	}

	return reminders, nil
}

// FindPending は送信待ちリマインダーを取得（未送信かつ送信時刻が現在時刻より前）
func (r *reminderRepository) FindPending(ctx context.Context) ([]domain.Reminder, error) {
	query := `
		SELECT id, todo_id, remind_at, is_sent, created_at
		FROM reminders
		WHERE is_sent = FALSE AND remind_at <= NOW()
		ORDER BY remind_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("送信待ちリマインダーの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var reminders []domain.Reminder
	for rows.Next() {
		var reminder domain.Reminder
		err := rows.Scan(&reminder.ID, &reminder.TodoID, &reminder.RemindAt, &reminder.IsSent, &reminder.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("リマインダーのスキャンに失敗しました: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("送信待ちリマインダーの取得に失敗しました: %w", err)
	}

	return reminders, nil
}

// MarkAsSent はリマインダーを送信済みにする
func (r *reminderRepository) MarkAsSent(ctx context.Context, reminderID int) error {
	query := `UPDATE reminders SET is_sent = TRUE WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, reminderID)
	if err != nil {
		return fmt.Errorf("送信済みの更新に失敗しました: %w", err)
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

// Delete はリマインダーを削除
func (r *reminderRepository) Delete(ctx context.Context, reminderID int) error {
	query := `DELETE FROM reminders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, reminderID)
	if err != nil {
		return fmt.Errorf("リマインダーの削除に失敗しました: %w", err)
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
