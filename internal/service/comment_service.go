// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package service

import (
	"context"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
)

// CommentService はコメント関連のビジネスロジックを提供
type CommentService struct {
	commentRepo repository.CommentRepository
}

// NewCommentService はCommentServiceの新しいインスタンスを作成
func NewCommentService(commentRepo repository.CommentRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
	}
}

// CreateComment はコメントを作成
func (s *CommentService) CreateComment(ctx context.Context, todoID int, input domain.CreateCommentInput, userID int) (domain.Comment, error) {
	return s.commentRepo.Create(ctx, todoID, input, userID)
}

// GetCommentsByTodoID は指定されたTODOの全てのコメントを取得
func (s *CommentService) GetCommentsByTodoID(ctx context.Context, todoID, userID int) ([]domain.Comment, error) {
	return s.commentRepo.FindByTodoID(ctx, todoID, userID)
}

// GetComment は指定されたIDのコメントを取得
func (s *CommentService) GetComment(ctx context.Context, commentID, userID int) (domain.Comment, error) {
	return s.commentRepo.FindByID(ctx, commentID, userID)
}

// UpdateComment はコメントを更新
func (s *CommentService) UpdateComment(ctx context.Context, commentID int, input domain.UpdateCommentInput, userID int) (domain.Comment, error) {
	return s.commentRepo.Update(ctx, commentID, input, userID)
}

// DeleteComment はコメントを削除
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	return s.commentRepo.Delete(ctx, commentID, userID)
}
