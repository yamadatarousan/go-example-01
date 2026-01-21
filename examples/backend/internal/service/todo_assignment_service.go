// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/backend/internal/domain" → "gin-quickstart/backend/internal/domain"

package service

import (
	"context"
	"errors"

	"gin-quickstart/examples/backend/internal/domain"
	"gin-quickstart/examples/backend/internal/repository"
)

// TodoAssignmentService はTODO担当者関連のビジネスロジックを提供
type TodoAssignmentService struct {
	assignmentRepo repository.TodoAssignmentRepository
	todoRepo       repository.TodoRepository
	projectRepo    repository.ProjectRepository
}

// NewTodoAssignmentService はTodoAssignmentServiceの新しいインスタンスを作成
func NewTodoAssignmentService(assignmentRepo repository.TodoAssignmentRepository, todoRepo repository.TodoRepository, projectRepo repository.ProjectRepository) *TodoAssignmentService {
	return &TodoAssignmentService{
		assignmentRepo: assignmentRepo,
		todoRepo:       todoRepo,
		projectRepo:    projectRepo,
	}
}

// AssignUser はTODOにユーザーを担当者として割り当て
func (s *TodoAssignmentService) AssignUser(ctx context.Context, todoID int, input domain.AssignUserInput, requesterID int) (domain.TodoAssignment, error) {
	// TODOへのアクセス権を確認
	todo, err := s.checkTodoAccessAndGetTodo(ctx, todoID, requesterID)
	if err != nil {
		return domain.TodoAssignment{}, err
	}

	// プロジェクトTODOの場合、割り当て先ユーザーがプロジェクトメンバーであることを確認
	if todo.ProjectID != nil {
		isMember, err := s.projectRepo.IsMember(ctx, *todo.ProjectID, input.UserID)
		if err != nil {
			return domain.TodoAssignment{}, err
		}
		if !isMember {
			return domain.TodoAssignment{}, errors.New("assigned user must be a project member")
		}
	}

	return s.assignmentRepo.AssignUser(ctx, todoID, input, requesterID)
}

// UnassignUser はTODOから担当者を解除
func (s *TodoAssignmentService) UnassignUser(ctx context.Context, todoID, userID, requesterID int) error {
	// TODOへのアクセス権を確認
	if _, err := s.checkTodoAccessAndGetTodo(ctx, todoID, requesterID); err != nil {
		return err
	}

	return s.assignmentRepo.UnassignUser(ctx, todoID, userID, requesterID)
}

// GetAssignments は指定されたTODOの全ての担当者を取得
func (s *TodoAssignmentService) GetAssignments(ctx context.Context, todoID, requesterID int) ([]domain.TodoAssignment, error) {
	// TODOへのアクセス権を確認
	if _, err := s.checkTodoAccessAndGetTodo(ctx, todoID, requesterID); err != nil {
		return nil, err
	}

	return s.assignmentRepo.GetAssignments(ctx, todoID, requesterID)
}

// checkTodoAccessAndGetTodo はTODOへのアクセス権を確認してTODOを返す（ビジネスロジック）
//
// アクセス権の判定ルール：
// 1. TODOの所有者である
// 2. プロジェクトTODOで、かつプロジェクトメンバーである
func (s *TodoAssignmentService) checkTodoAccessAndGetTodo(ctx context.Context, todoID, userID int) (domain.Todo, error) {
	// TODOを取得してユーザーの所有かチェック
	todo, err := s.todoRepo.FindByID(ctx, todoID, userID)
	if err == nil {
		// ユーザーの所有TODOならアクセス可能
		return todo, nil
	}

	// ユーザーの所有でない場合、プロジェクトメンバーとしてアクセス可能かチェック
	// システム権限でTODO情報を取得
	todoInfo, err := s.todoRepo.FindByID(ctx, todoID, 0)
	if err != nil {
		return domain.Todo{}, errors.New("todo not found or access denied")
	}

	// プロジェクトTODOでない場合はアクセス拒否
	if todoInfo.ProjectID == nil {
		return domain.Todo{}, errors.New("access denied: not the todo owner")
	}

	// プロジェクトメンバーであることを確認
	isMember, err := s.projectRepo.IsMember(ctx, *todoInfo.ProjectID, userID)
	if err != nil || !isMember {
		return domain.Todo{}, errors.New("access denied: not a project member")
	}

	return todoInfo, nil
}
