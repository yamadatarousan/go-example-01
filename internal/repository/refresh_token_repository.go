package repository

import (
	"context"
	"database/sql"

	"gin-quickstart/examples/internal/domain"
)

// RefreshTokenRepository はリフレッシュトークンのデータアクセス層インターフェース
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token domain.RefreshToken) (domain.RefreshToken, error)
	FindRefreshTokenByToken(ctx context.Context, token string) (domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
}

// PostgresRefreshTokenRepository はRefreshTokenRepositoryのPostgreSQL実装
type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

// NewPostgresRefreshTokenRepository はPostgresRefreshTokenRepositoryの新しいインスタンスを作成
func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

// CreateRefreshToken は新しいリフレッシュトークンを作成
func (r *PostgresRefreshTokenRepository) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) (domain.RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, token, user_id, expires_at, created_at, revoked
	`

	var createdToken domain.RefreshToken
	err := r.db.QueryRowContext(
		ctx,
		query,
		token.Token,
		token.UserID,
		token.ExpiresAt,
	).Scan(
		&createdToken.ID,
		&createdToken.Token,
		&createdToken.UserID,
		&createdToken.ExpiresAt,
		&createdToken.CreatedAt,
		&createdToken.Revoked,
	)

	if err != nil {
		return domain.RefreshToken{}, err
	}

	return createdToken, nil
}

// FindRefreshTokenByToken はトークン文字列でリフレッシュトークンを検索
func (r *PostgresRefreshTokenRepository) FindRefreshTokenByToken(ctx context.Context, token string) (domain.RefreshToken, error) {
	query := `
		SELECT id, token, user_id, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1 AND revoked = FALSE
	`

	var refreshToken domain.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.Token,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
		&refreshToken.Revoked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.RefreshToken{}, ErrNotFound
		}
		return domain.RefreshToken{}, err
	}

	return refreshToken, nil
}

// RevokeRefreshToken はリフレッシュトークンを無効化
func (r *PostgresRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token = $1
	`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteExpiredTokens は有効期限切れのトークンを削除
func (r *PostgresRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}
