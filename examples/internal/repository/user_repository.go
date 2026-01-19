package repository

import (
	"context"
	"database/sql"

	"gin-quickstart/examples/internal/domain"
)

// userRepository はUserRepositoryインターフェースの実装
type userRepository struct {
	db *sql.DB
}

// NewUserRepository はUserRepositoryの新しいインスタンスを作成
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser は新しいユーザーを作成
func (r *userRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, role",
		user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

// FindUserByEmail はメールアドレスでユーザーを検索
func (r *userRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, password_hash, created_at, role FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

// FindUserByID はユーザーIDでユーザーを検索
func (r *userRepository) FindUserByID(ctx context.Context, userID int) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, password_hash, created_at, role FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

// FindAllUsers は全てのユーザーを取得
func (r *userRepository) FindAllUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, email, created_at, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
