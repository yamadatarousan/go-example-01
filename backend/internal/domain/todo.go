package domain

import (
	"errors"
	"time"
)

// Todo represents a todo item in the domain model
type Todo struct {
	ID           int        `json:"id"`
	Name         string     `json:"name" binding:"required"`
	Description  *string    `json:"description"`                                            // 詳細説明（NULL可）
	Status       string     `json:"status" binding:"omitempty,oneof=todo in_progress done"` // ステータス
	Priority     string     `json:"priority" binding:"omitempty,oneof=low medium high"`     // 優先度
	DueDate      *time.Time `json:"due_date"`                                               // 期限（NULL可）
	UserID       int        `json:"user_id"`
	CategoryID   *int       `json:"category_id"`    // カテゴリーID（NULL可）
	ParentTodoID *int       `json:"parent_todo_id"` // 親TODO（サブタスク用、NULL可）
	ProjectID    *int       `json:"project_id"`     // プロジェクトID（NULL可、Phase 5で追加）
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// リレーション（JOIN時に使用）
	Category *Category `json:"category,omitempty"` // カテゴリー情報
	Tags     []Tag     `json:"tags,omitempty"`     // タグ一覧
}

// CreateTodoInput はTODO作成時の入力
type CreateTodoInput struct {
	Name         string     `json:"name" binding:"required"`
	Description  *string    `json:"description"`
	Status       string     `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority     string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"due_date"`
	CategoryID   *int       `json:"category_id"`
	ParentTodoID *int       `json:"parent_todo_id"`
	ProjectID    *int       `json:"project_id"` // プロジェクトID（Phase 5で追加）
	TagIDs       []int      `json:"tag_ids"`    // タグIDの配列
}

// UpdateTodoInput はTODO更新時の入力
type UpdateTodoInput struct {
	Name         *string    `json:"name"`
	Description  *string    `json:"description"`
	Status       *string    `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority     *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"due_date"`
	CategoryID   *int       `json:"category_id"`
	ParentTodoID *int       `json:"parent_todo_id"`
	ProjectID    *int       `json:"project_id"` // プロジェクトID（Phase 5で追加）
	TagIDs       *[]int     `json:"tag_ids"`    // タグIDの配列（NULL可）
}

// ============================================================================
// Phase 3で追加された型
// ============================================================================

// SearchFilters は検索条件を表す
type SearchFilters struct {
	Status     *string    `form:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority   *string    `form:"priority" binding:"omitempty,oneof=low medium high"`
	CategoryID *int       `form:"category_id"`
	TagIDs     []int      `form:"tag_ids"`  // タグID配列
	Search     string     `form:"search"`   // 全文検索キーワード
	DueFrom    *time.Time `form:"due_from"` // 期限開始日
	DueTo      *time.Time `form:"due_to"`   // 期限終了日
	Sort       string     `form:"sort"`     // ソート項目（due_date, priority, created_at）
	Order      string     `form:"order"`    // ソート順（asc, desc）
	Page       int        `form:"page"`     // ページ番号
	Limit      int        `form:"limit"`    // 1ページあたりの件数
}

// SearchResult は検索結果を表す
type SearchResult struct {
	Todos      []Todo `json:"todos"`       // TODO一覧
	Total      int    `json:"total"`       // 総件数
	Page       int    `json:"page"`        // 現在のページ
	Limit      int    `json:"limit"`       // 1ページあたりの件数
	TotalPages int    `json:"total_pages"` // 総ページ数
}

// Statistics はTODO統計情報を表す
type Statistics struct {
	TotalCount       int            `json:"total_count"`         // 総TODO数
	StatusCounts     map[string]int `json:"status_counts"`       // ステータス別カウント
	PriorityCounts   map[string]int `json:"priority_counts"`     // 優先度別カウント
	OverdueCount     int            `json:"overdue_count"`       // 期限切れ数
	DueTodayCount    int            `json:"due_today_count"`     // 今日期限数
	DueThisWeekCount int            `json:"due_this_week_count"` // 今週期限数
}

// ============================================================================
// ドメインエラー
// ============================================================================

var (
	// ErrAlreadyCompleted はTODOが既に完了済みの場合のエラー
	ErrAlreadyCompleted = errors.New("todo is already completed")
	// ErrNotCompleted はTODOが完了していない場合のエラー
	ErrNotCompleted = errors.New("todo is not completed")
)

// ============================================================================
// ドメインロジック（ビジネスルール）
// ============================================================================

// Complete はTODOを完了状態にする
//
// ビジネスルール:
// - 既に完了済みの場合はエラー
// - ステータスを "done" に変更
func (t *Todo) Complete() error {
	if t.Status == "done" {
		return ErrAlreadyCompleted
	}
	t.Status = "done"
	return nil
}

// Reopen はTODOを再開する（未完了状態に戻す）
//
// ビジネスルール:
// - 完了済みでない場合はエラー
// - ステータスを "todo" に変更
func (t *Todo) Reopen() error {
	if t.Status != "done" {
		return ErrNotCompleted
	}
	t.Status = "todo"
	return nil
}
