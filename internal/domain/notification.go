package domain

import "time"

// Notification は通知を表すドメインモデル
type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	TodoID    *int      `json:"todo_id"` // NULL可（TODOに紐づかない通知もある）
	Type      string    `json:"type"`    // 'deadline_reminder', 'todo_assigned', 'todo_completed'
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateNotificationInput は通知作成時の入力
type CreateNotificationInput struct {
	UserID  int    `json:"user_id" binding:"required"`
	TodoID  *int   `json:"todo_id"`
	Type    string `json:"type" binding:"required,oneof=deadline_reminder todo_assigned todo_completed"`
	Message string `json:"message" binding:"required"`
}
