package main

import (
	"context"
	"database/sql"
	"fmt"
)

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) FindAll(userID int) ([]Todo, error) {
	rows, err := r.db.Query("SELECT id, name, user_id FROM todos WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Name, &todo.UserID)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (r *TodoRepository) Create(todo Todo) (Todo, error) {
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

func (r *TodoRepository) CreateUser(user User) (User, error) {
	err := r.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, role",
		user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *TodoRepository) FindUserByEmail(email string) (User, error) {
	var user User
	err := r.db.QueryRow("SELECT id, email, password_hash, created_at, role FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

// execTxはトランザクションを実行するためのヘルパー関数です
// トランザクションを開始し、渡された関数(fn)を実行します。
// fnがエラーを返した場合、トランザクションはロールバックされます。
// エラーがなければ、トランザクションはコミットされます。
func (r *TodoRepository) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
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

// createTodoInTxはトランザクション内でTODOと監査ログを作成します。
func (r *TodoRepository) createTodoInTx(tx *sql.Tx, todo Todo) (Todo, error) {
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

// CreateTodoWithAuditはトランザクションを使用してTODOと監査ログを作成します。
func (r *TodoRepository) CreateTodoWithAudit(ctx context.Context, todo Todo) (Todo, error) {
	var createdTodo Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		createdTodo, err = r.createTodoInTx(tx, todo)
		return err
	})

	return createdTodo, err
}

func (r *TodoRepository) FindAllUsers() ([]User, error) {
	rows, err := r.db.Query("SELECT id, email, created_at, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// FindByIDは指定されたIDのTODOを取得します
func (r *TodoRepository) FindByID(todoID, userID int) (Todo, error) {
	var todo Todo
	err := r.db.QueryRow(
		"SELECT id, name, user_id FROM todos WHERE id = $1 AND user_id = $2",
		todoID, userID,
	).Scan(&todo.ID, &todo.Name, &todo.UserID)
	if err != nil {
		return todo, err
	}
	return todo, nil
}

// updateTodoInTxはトランザクション内でTODOを更新し、監査ログを作成します
func (r *TodoRepository) updateTodoInTx(tx *sql.Tx, todo Todo) (Todo, error) {
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

// UpdateTodoWithAuditはトランザクションを使用してTODOを更新します
func (r *TodoRepository) UpdateTodoWithAudit(ctx context.Context, todo Todo) (Todo, error) {
	var updatedTodo Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		updatedTodo, err = r.updateTodoInTx(tx, todo)
		return err
	})

	return updatedTodo, err
}
