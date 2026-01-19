package domain

import "time"

// RefreshToken はリフレッシュトークンのドメインモデル
type RefreshToken struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}
