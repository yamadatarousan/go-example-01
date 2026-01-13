package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gin-quickstart/internal/domain"
)

// todoRepository はTodoRepositoryインターフェースの実装
type todoRepository struct {
	db *sql.DB
}

// NewTodoRepository はTodoRepositoryの新しいインスタンスを作成
func NewTodoRepository(db *sql.DB) TodoRepository {
	return &todoRepository{db: db}
}

// FindAll は指定されたユーザーの全てのTODOを取得
func (r *todoRepository) FindAll(userID int) ([]domain.Todo, error) {
	rows, err := r.db.Query("SELECT id, name, user_id FROM todos WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		err := rows.Scan(&todo.ID, &todo.Name, &todo.UserID)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// FindByID は指定されたIDのTODOを取得
func (r *todoRepository) FindByID(todoID, userID int) (domain.Todo, error) {
	var todo domain.Todo
	err := r.db.QueryRow(
		"SELECT id, name, user_id FROM todos WHERE id = $1 AND user_id = $2",
		todoID, userID,
	).Scan(&todo.ID, &todo.Name, &todo.UserID)
	if err != nil {
		return todo, err
	}
	return todo, nil
}

// Create はTODOを作成
func (r *todoRepository) Create(todo domain.Todo) (domain.Todo, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO todos (name, user_id) VALUES ($1, $2) RETURNING id",
		todo.Name, todo.UserID,
	).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id
	return todo, nil
}

// execTx はトランザクションを実行するためのヘルパー関数
// トランザクションを開始し、渡された関数(fn)を実行します。
// fnがエラーを返した場合、トランザクションはロールバックされます。
// エラーがなければ、トランザクションはコミットされます。
func (r *todoRepository) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		// エラーが発生した場合、ロールバックを試みる
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// createTodoInTx はトランザクション内でTODOと監査ログを作成
func (r *todoRepository) createTodoInTx(tx *sql.Tx, todo domain.Todo) (domain.Todo, error) {
	// 1. todosテーブルに新しいTODOを挿入し、IDを取得
	var id int
	err := tx.QueryRow("INSERT INTO todos (name, user_id) VALUES ($1, $2) RETURNING id", todo.Name, todo.UserID).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id

	// 2. todo_audit_logsテーブルに監査ログを挿入
	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", id, "create")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

// CreateTodoWithAudit はトランザクションを使用してTODOと監査ログを作成
func (r *todoRepository) CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	var createdTodo domain.Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		createdTodo, err = r.createTodoInTx(tx, todo)
		return err
	})

	return createdTodo, err
}

// updateTodoInTx はトランザクション内でTODOを更新し、監査ログを作成
func (r *todoRepository) updateTodoInTx(tx *sql.Tx, todo domain.Todo) (domain.Todo, error) {
	// 1. todosテーブルのレコードを更新
	result, err := tx.Exec(
		"UPDATE todos SET name = $1 WHERE id = $2 AND user_id = $3",
		todo.Name, todo.ID, todo.UserID,
	)
	if err != nil {
		return todo, err
	}

	// 更新された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return todo, err
	}
	if rowsAffected == 0 {
		return todo, sql.ErrNoRows
	}

	// 2. todo_audit_logsテーブルに監査ログを挿入
	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", todo.ID, "update")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

// UpdateTodoWithAudit はトランザクションを使用してTODOを更新
func (r *todoRepository) UpdateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	var updatedTodo domain.Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		updatedTodo, err = r.updateTodoInTx(tx, todo)
		return err
	})

	return updatedTodo, err
}

// deleteTodoInTx はトランザクション内でTODOを削除し、監査ログを作成
func (r *todoRepository) deleteTodoInTx(tx *sql.Tx, todoID, userID int) error {
	// 1. 監査ログを先に挿入（TODOが削除される前に）
	_, err := tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", todoID, "delete")
	if err != nil {
		return err
	}

	// 2. todosテーブルからレコードを削除
	result, err := tx.Exec("DELETE FROM todos WHERE id = $1 AND user_id = $2", todoID, userID)
	if err != nil {
		return err
	}

	// 削除された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteTodoWithAudit はトランザクションを使用してTODOを削除
func (r *todoRepository) DeleteTodoWithAudit(ctx context.Context, todoID, userID int) error {
	return r.execTx(ctx, func(tx *sql.Tx) error {
		return r.deleteTodoInTx(tx, todoID, userID)
	})
}
