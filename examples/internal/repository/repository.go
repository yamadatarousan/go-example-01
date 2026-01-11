package repository

import (
	"context"

	"gin-quickstart/examples/internal/domain"
)

// TodoRepository はTODOデータアクセスのインターフェースを定義
type TodoRepository interface {
	FindAll(userID int) ([]domain.Todo, error)
	FindByID(todoID, userID int) (domain.Todo, error)
	Create(todo domain.Todo) (domain.Todo, error)
	CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	UpdateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	DeleteTodoWithAudit(ctx context.Context, todoID, userID int) error
}

// UserRepository はユーザーデータアクセスのインターフェースを定義
type UserRepository interface {
	CreateUser(user domain.User) (domain.User, error)
	FindUserByEmail(email string) (domain.User, error)
	FindAllUsers() ([]domain.User, error)
}
