// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package service

import (
	"context"
	"errors"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
)

// CommentService はコメント関連のビジネスロジックを提供
type CommentService struct {
	commentRepo repository.CommentRepository
	todoRepo    repository.TodoRepository
	projectRepo repository.ProjectRepository
}

// NewCommentService はCommentServiceの新しいインスタンスを作成
func NewCommentService(commentRepo repository.CommentRepository, todoRepo repository.TodoRepository, projectRepo repository.ProjectRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		todoRepo:    todoRepo,
		projectRepo: projectRepo,
	}
}

// CreateComment はコメントを作成
func (s *CommentService) CreateComment(ctx context.Context, todoID int, input domain.CreateCommentInput, userID int) (domain.Comment, error) {
	// TODOへのアクセス権を確認
	if err := s.checkTodoAccess(ctx, todoID, userID); err != nil {
		return domain.Comment{}, err
	}

	return s.commentRepo.Create(ctx, todoID, input, userID)
}

// GetCommentsByTodoID は指定されたTODOの全てのコメントを取得
func (s *CommentService) GetCommentsByTodoID(ctx context.Context, todoID, userID int) ([]domain.Comment, error) {
	// TODOへのアクセス権を確認
	if err := s.checkTodoAccess(ctx, todoID, userID); err != nil {
		return nil, err
	}

	return s.commentRepo.FindByTodoID(ctx, todoID, userID)
}

// GetComment は指定されたIDのコメントを取得
func (s *CommentService) GetComment(ctx context.Context, commentID, userID int) (domain.Comment, error) {
	// コメントを取得
	comment, err := s.commentRepo.FindByID(ctx, commentID, userID)
	if err != nil {
		return domain.Comment{}, err
	}

	// TODOへのアクセス権を確認
	if err := s.checkTodoAccess(ctx, comment.TodoID, userID); err != nil {
		return domain.Comment{}, err
	}

	return comment, nil
}

// UpdateComment はコメントを更新
func (s *CommentService) UpdateComment(ctx context.Context, commentID int, input domain.UpdateCommentInput, userID int) (domain.Comment, error) {
	return s.commentRepo.Update(ctx, commentID, input, userID)
}

// DeleteComment はコメントを削除
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	return s.commentRepo.Delete(ctx, commentID, userID)
}

// checkTodoAccess はTODOへのアクセス権を確認（ビジネスロジック）
//
// アクセス権の判定ルール：
// 1. TODOの所有者である
// 2. プロジェクトTODOで、かつプロジェクトメンバーである
func (s *CommentService) checkTodoAccess(ctx context.Context, todoID, userID int) error {
	// TODOを取得してユーザーの所有かチェック
	_, err := s.todoRepo.FindByID(ctx, todoID, userID)
	if err == nil {
		// ユーザーの所有TODOならアクセス可能
		return nil
	}

	// ユーザーの所有でない場合、プロジェクトメンバーとしてアクセス可能かチェック
	// システム権限でTODO情報を取得
	todoInfo, err := s.todoRepo.FindByID(ctx, todoID, 0)
	if err != nil {
		return errors.New("todo not found or access denied")
	}

	// プロジェクトTODOでない場合はアクセス拒否
	if todoInfo.ProjectID == nil {
		return errors.New("access denied: not the todo owner")
	}

	// プロジェクトメンバーであることを確認
	isMember, err := s.projectRepo.IsMember(ctx, *todoInfo.ProjectID, userID)
	if err != nil || !isMember {
		return errors.New("access denied: not a project member")
	}

	return nil
}
