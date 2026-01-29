
package domain

import "time"

// TodoAssignment はTODO担当者を表すドメインモデル
type TodoAssignment struct {
	TodoID     int       `json:"todo_id"`
	UserID     int       `json:"user_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

// AssignUserInput はTODO担当者割り当て時の入力
type AssignUserInput struct {
	UserID int `json:"user_id" binding:"required"`
}
