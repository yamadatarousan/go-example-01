package service

import (
	"context"

	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/repository"
)

// TodoService はTODO関連のビジネスロジックを提供
type TodoService struct {
	todoRepo repository.TodoRepository
}

// NewTodoService はTodoServiceの新しいインスタンスを作成
func NewTodoService(todoRepo repository.TodoRepository) *TodoService {
	return &TodoService{
		todoRepo: todoRepo,
	}
}

// GetTodos は指定されたユーザーの全てのTODOを取得
func (s *TodoService) GetTodos(ctx context.Context, userID int) ([]domain.Todo, error) {
	return s.todoRepo.FindAll(ctx, userID)
}

// GetTodo は指定されたIDのTODOを取得
func (s *TodoService) GetTodo(ctx context.Context, todoID, userID int) (domain.Todo, error) {
	return s.todoRepo.FindByID(ctx, todoID, userID)
}

// CreateTodo はTODOを作成（監査ログ付き）
func (s *TodoService) CreateTodo(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	return s.todoRepo.CreateTodoWithAudit(ctx, todo)
}

// UpdateTodo はTODOを更新（監査ログ付き）
func (s *TodoService) UpdateTodo(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	return s.todoRepo.UpdateTodoWithAudit(ctx, todo)
}

// DeleteTodo はTODOを削除（監査ログ付き）
func (s *TodoService) DeleteTodo(ctx context.Context, todoID, userID int) error {
	return s.todoRepo.DeleteTodoWithAudit(ctx, todoID, userID)
}

// ============================================================================
// Phase 2で追加されたメソッド
// ============================================================================

// CompleteTodo はTODOを完了状態にする
func (s *TodoService) CompleteTodo(ctx context.Context, todoID, userID int) error {
	return s.todoRepo.UpdateStatus(ctx, todoID, userID, "done")
}

// ReopenTodo はTODOを再開する（todoステータスに戻す）
func (s *TodoService) ReopenTodo(ctx context.Context, todoID, userID int) error {
	return s.todoRepo.UpdateStatus(ctx, todoID, userID, "todo")
}

// GetOverdueTodos は期限切れのTODOを取得
func (s *TodoService) GetOverdueTodos(ctx context.Context, userID int) ([]domain.Todo, error) {
	return s.todoRepo.FindOverdue(ctx, userID)
}

// GetTodayTodos は今日が期限のTODOを取得
func (s *TodoService) GetTodayTodos(ctx context.Context, userID int) ([]domain.Todo, error) {
	return s.todoRepo.FindToday(ctx, userID)
}

// GetThisWeekTodos は今週が期限のTODOを取得
func (s *TodoService) GetThisWeekTodos(ctx context.Context, userID int) ([]domain.Todo, error) {
	return s.todoRepo.FindThisWeek(ctx, userID)
}

// ============================================================================
// Phase 3で追加されたメソッド
// ============================================================================

// SearchTodos は高度な検索・フィルタリング機能を提供
func (s *TodoService) SearchTodos(ctx context.Context, userID int, filters domain.SearchFilters) (domain.SearchResult, error) {
	return s.todoRepo.Search(ctx, userID, filters)
}

// GetStatistics はTODO統計情報を取得
func (s *TodoService) GetStatistics(ctx context.Context, userID int) (domain.Statistics, error) {
	return s.todoRepo.GetStatistics(ctx, userID)
}
