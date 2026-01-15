// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package repository

import (
	"context"
	"database/sql"
	"errors"

	"gin-quickstart/examples/internal/domain"
)

// TodoAssignmentRepository はTODO担当者のリポジトリインターフェース
type TodoAssignmentRepository interface {
	AssignUser(ctx context.Context, todoID int, input domain.AssignUserInput, requesterID int) (domain.TodoAssignment, error)
	UnassignUser(ctx context.Context, todoID, userID, requesterID int) error
	GetAssignments(ctx context.Context, todoID, requesterID int) ([]domain.TodoAssignment, error)
}

// todoAssignmentRepository はTodoAssignmentRepositoryインターフェースの実装
type todoAssignmentRepository struct {
	db          *sql.DB
	todoRepo    TodoRepository
	projectRepo ProjectRepository
}

// NewTodoAssignmentRepository はTodoAssignmentRepositoryの新しいインスタンスを作成
func NewTodoAssignmentRepository(db *sql.DB, todoRepo TodoRepository, projectRepo ProjectRepository) TodoAssignmentRepository {
	return &todoAssignmentRepository{
		db:          db,
		todoRepo:    todoRepo,
		projectRepo: projectRepo,
	}
}

// AssignUser はTODOにユーザーを担当者として割り当て
func (r *todoAssignmentRepository) AssignUser(ctx context.Context, todoID int, input domain.AssignUserInput, requesterID int) (domain.TodoAssignment, error) {
	var assignment domain.TodoAssignment

	// TODOへのアクセス権を確認
	todo, err := r.checkTodoAccessAndGetTodo(ctx, todoID, requesterID)
	if err != nil {
		return assignment, err
	}

	// プロジェクトTODOの場合、割り当て先ユーザーがプロジェクトメンバーであることを確認
	if todo.ProjectID != nil {
		isMember, err := r.projectRepo.IsMember(ctx, *todo.ProjectID, input.UserID)
		if err != nil {
			return assignment, err
		}
		if !isMember {
			return assignment, errors.New("assigned user must be a project member")
		}
	}

	// 担当者を割り当て
	err = r.db.QueryRowContext(ctx, `
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
	// TODOへのアクセス権を確認
	if _, err := r.checkTodoAccessAndGetTodo(ctx, todoID, requesterID); err != nil {
		return err
	}

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
	// TODOへのアクセス権を確認
	if _, err := r.checkTodoAccessAndGetTodo(ctx, todoID, requesterID); err != nil {
		return nil, err
	}

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

// checkTodoAccessAndGetTodo はTODOへのアクセス権を確認してTODOを返す
func (r *todoAssignmentRepository) checkTodoAccessAndGetTodo(ctx context.Context, todoID, userID int) (domain.Todo, error) {
	// TODOを取得
	todo, err := r.todoRepo.FindByID(ctx, todoID, userID)
	if err != nil {
		// ユーザーの所有TODOでない場合、プロジェクトメンバーとしてアクセス可能かチェック
		// まずTODO情報を取得（システム権限で）
		todoInfo, sysErr := r.todoRepo.FindByID(ctx, todoID, 0) // userID=0でシステム権限
		if sysErr != nil {
			return domain.Todo{}, errors.New("todo not found or access denied")
		}

		// プロジェクトTODOであることを確認
		if todoInfo.ProjectID == nil {
			return domain.Todo{}, errors.New("access denied: not the todo owner")
		}

		// プロジェクトメンバーであることを確認
		isMember, memberErr := r.projectRepo.IsMember(ctx, *todoInfo.ProjectID, userID)
		if memberErr != nil || !isMember {
			return domain.Todo{}, errors.New("access denied: not a project member")
		}

		// プロジェクトメンバーならアクセス可能
		return todoInfo, nil
	}

	return todo, nil
}
