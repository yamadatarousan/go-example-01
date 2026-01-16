// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package repository

import (
	"context"
	"database/sql"
	"errors"

	"gin-quickstart/internal/domain"
)

// commentRepository はCommentRepositoryインターフェースの実装
type commentRepository struct {
	db          *sql.DB
	todoRepo    TodoRepository
	projectRepo ProjectRepository
}

// NewCommentRepository はCommentRepositoryの新しいインスタンスを作成
func NewCommentRepository(db *sql.DB, todoRepo TodoRepository, projectRepo ProjectRepository) CommentRepository {
	return &commentRepository{
		db:          db,
		todoRepo:    todoRepo,
		projectRepo: projectRepo,
	}
}

// Create は新しいコメントを作成
func (r *commentRepository) Create(ctx context.Context, todoID int, input domain.CreateCommentInput, userID int) (domain.Comment, error) {
	var comment domain.Comment

	// TODOへのアクセス権を確認
	if err := r.checkTodoAccess(ctx, todoID, userID); err != nil {
		return comment, err
	}

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
	// TODOへのアクセス権を確認
	if err := r.checkTodoAccess(ctx, todoID, userID); err != nil {
		return nil, err
	}

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

	// コメントを取得
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

	// TODOへのアクセス権を確認
	if err := r.checkTodoAccess(ctx, comment.TodoID, userID); err != nil {
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

// checkTodoAccess はTODOへのアクセス権を確認
func (r *commentRepository) checkTodoAccess(ctx context.Context, todoID, userID int) error {
	// TODOを取得してアクセス権を確認
	_, err := r.todoRepo.FindByID(ctx, todoID, userID)
	if err != nil {
		// ユーザーの所有TODOでない場合、プロジェクトメンバーとしてアクセス可能かチェック
		// まずTODO情報を取得（システム権限で）
		todoInfo, sysErr := r.todoRepo.FindByID(ctx, todoID, 0) // userID=0でシステム権限
		if sysErr != nil {
			return errors.New("todo not found or access denied")
		}

		// プロジェクトTODOであることを確認
		if todoInfo.ProjectID == nil {
			return errors.New("access denied: not the todo owner")
		}

		// プロジェクトメンバーであることを確認
		isMember, memberErr := r.projectRepo.IsMember(ctx, *todoInfo.ProjectID, userID)
		if memberErr != nil || !isMember {
			return errors.New("access denied: not a project member")
		}

		// プロジェクトメンバーならアクセス可能
		return nil
	}

	// TODOが見つかり、ユーザーの所有である場合
	return nil
}
