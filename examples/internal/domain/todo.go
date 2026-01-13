package domain

import (
	"time"
)

// Todo represents a todo item in the domain model
type Todo struct {
	ID          int        `json:"id"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`                              // 詳細説明（NULL可）
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done"` // ステータス
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`     // 優先度
	DueDate     *time.Time `json:"due_date"`                                 // 期限（NULL可）
	UserID      int        `json:"user_id"`
	CategoryID  *int       `json:"category_id"`                              // カテゴリーID（NULL可）
	ParentTodoID *int      `json:"parent_todo_id"`                           // 親TODO（サブタスク用、NULL可）
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// リレーション（JOIN時に使用）
	Category *Category `json:"category,omitempty"` // カテゴリー情報
	Tags     []Tag     `json:"tags,omitempty"`     // タグ一覧
}

// CreateTodoInput はTODO作成時の入力
type CreateTodoInput struct {
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date"`
	CategoryID  *int       `json:"category_id"`
	ParentTodoID *int      `json:"parent_todo_id"`
	TagIDs      []int      `json:"tag_ids"` // タグIDの配列
}

// UpdateTodoInput はTODO更新時の入力
type UpdateTodoInput struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date"`
	CategoryID  *int       `json:"category_id"`
	ParentTodoID *int      `json:"parent_todo_id"`
	TagIDs      *[]int     `json:"tag_ids"` // タグIDの配列（NULL可）
}
