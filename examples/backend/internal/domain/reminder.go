package domain

import "time"

// Reminder はリマインダーを表すドメインモデル
type Reminder struct {
	ID        int       `json:"id"`
	TodoID    int       `json:"todo_id"`
	RemindAt  time.Time `json:"remind_at"`
	IsSent    bool      `json:"is_sent"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateReminderInput はリマインダー作成時の入力
type CreateReminderInput struct {
	TodoID   int       `json:"todo_id" binding:"required"`
	RemindAt time.Time `json:"remind_at" binding:"required"`
}
