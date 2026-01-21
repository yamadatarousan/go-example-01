// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/backend/internal/domain" → "gin-quickstart/backend/internal/domain"

package repository

import (
	"context"
	"database/sql"
	"errors"

	"gin-quickstart/backend/internal/domain"
)

// commentRepository はCommentRepositoryインターフェースの実装
type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository はCommentRepositoryの新しいインスタンスを作成
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}

// Create は新しいコメントを作成
func (r *commentRepository) Create(ctx context.Context, todoID int, input domain.CreateCommentInput, userID int) (domain.Comment, error) {
	var comment domain.Comment

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO comments (todo_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, todo_id, user_id, content, created_at, updated_at
	`, todoID, userID, input.Content).Scan(
		&comment.ID, &comment.TodoID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return comment, err
	}

	return comment, nil
}

// FindByTodoID は指定されたTODOの全てのコメントを取得
func (r *commentRepository) FindByTodoID(ctx context.Context, todoID, userID int) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, todo_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE todo_id = $1
		ORDER BY created_at ASC
	`, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var comment domain.Comment
		err := rows.Scan(&comment.ID, &comment.TodoID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// FindByID は指定されたIDのコメントを取得
func (r *commentRepository) FindByID(ctx context.Context, commentID, userID int) (domain.Comment, error) {
	var comment domain.Comment

	err := r.db.QueryRowContext(ctx, `
		SELECT id, todo_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE id = $1
	`, commentID).Scan(
		&comment.ID, &comment.TodoID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return comment, errors.New("comment not found")
		}
		return comment, err
	}

	return comment, nil
}

// Update はコメントを更新（作成者のみ可能）
func (r *commentRepository) Update(ctx context.Context, commentID int, input domain.UpdateCommentInput, userID int) (domain.Comment, error) {
	var comment domain.Comment

	// コメントの所有者であることを確認
	err := r.db.QueryRowContext(ctx, `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, todo_id, user_id, content, created_at, updated_at
	`, input.Content, commentID, userID).Scan(
		&comment.ID, &comment.TodoID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return comment, errors.New("comment not found or access denied")
		}
		return comment, err
	}

	return comment, nil
}

// Delete はコメントを削除（作成者のみ可能）
func (r *commentRepository) Delete(ctx context.Context, commentID, userID int) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM comments WHERE id = $1 AND user_id = $2
	`, commentID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("comment not found or access denied")
	}

	return nil
}
