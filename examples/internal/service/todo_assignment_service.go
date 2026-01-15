// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package service

import (
	"context"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
)

// TodoAssignmentService はTODO担当者関連のビジネスロジックを提供
type TodoAssignmentService struct {
	assignmentRepo repository.TodoAssignmentRepository
}

// NewTodoAssignmentService はTodoAssignmentServiceの新しいインスタンスを作成
func NewTodoAssignmentService(assignmentRepo repository.TodoAssignmentRepository) *TodoAssignmentService {
	return &TodoAssignmentService{
		assignmentRepo: assignmentRepo,
	}
}

// AssignUser はTODOにユーザーを担当者として割り当て
func (s *TodoAssignmentService) AssignUser(ctx context.Context, todoID int, input domain.AssignUserInput, requesterID int) (domain.TodoAssignment, error) {
	return s.assignmentRepo.AssignUser(ctx, todoID, input, requesterID)
}

// UnassignUser はTODOから担当者を解除
func (s *TodoAssignmentService) UnassignUser(ctx context.Context, todoID, userID, requesterID int) error {
	return s.assignmentRepo.UnassignUser(ctx, todoID, userID, requesterID)
}

// GetAssignments は指定されたTODOの全ての担当者を取得
func (s *TodoAssignmentService) GetAssignments(ctx context.Context, todoID, requesterID int) ([]domain.TodoAssignment, error) {
	return s.assignmentRepo.GetAssignments(ctx, todoID, requesterID)
}
