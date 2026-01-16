// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package repository

import (
	"context"
	"database/sql"
	"errors"

	"gin-quickstart/examples/internal/domain"
)

// todoAssignmentRepository はTodoAssignmentRepositoryインターフェースの実装
type todoAssignmentRepository struct {
	db *sql.DB
}

// NewTodoAssignmentRepository はTodoAssignmentRepositoryの新しいインスタンスを作成
func NewTodoAssignmentRepository(db *sql.DB) TodoAssignmentRepository {
	return &todoAssignmentRepository{
		db: db,
	}
}

// AssignUser はTODOにユーザーを担当者として割り当て
func (r *todoAssignmentRepository) AssignUser(ctx context.Context, todoID int, input domain.AssignUserInput, requesterID int) (domain.TodoAssignment, error) {
	var assignment domain.TodoAssignment

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO todo_assignments (todo_id, user_id)
		VALUES ($1, $2)
		RETURNING todo_id, user_id, assigned_at
	`, todoID, input.UserID).Scan(
		&assignment.TodoID, &assignment.UserID, &assignment.AssignedAt,
	)
	if err != nil {
		return assignment, err
	}

	return assignment, nil
}

// UnassignUser はTODOから担当者を解除
func (r *todoAssignmentRepository) UnassignUser(ctx context.Context, todoID, userID, requesterID int) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM todo_assignments WHERE todo_id = $1 AND user_id = $2
	`, todoID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("assignment not found")
	}

	return nil
}

// GetAssignments は指定されたTODOの全ての担当者を取得
func (r *todoAssignmentRepository) GetAssignments(ctx context.Context, todoID, requesterID int) ([]domain.TodoAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT todo_id, user_id, assigned_at
		FROM todo_assignments
		WHERE todo_id = $1
		ORDER BY assigned_at ASC
	`, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []domain.TodoAssignment
	for rows.Next() {
		var assignment domain.TodoAssignment
		err := rows.Scan(&assignment.TodoID, &assignment.UserID, &assignment.AssignedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}
