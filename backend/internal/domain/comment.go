// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/backend/internal/domain" → "gin-quickstart/backend/internal/domain"

package domain

import "time"

// Comment はTODOコメントを表すドメインモデル
type Comment struct {
	ID        int       `json:"id"`
	TodoID    int       `json:"todo_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCommentInput はコメント作成時の入力
type CreateCommentInput struct {
	Content string `json:"content" binding:"required"`
}

// UpdateCommentInput はコメント更新時の入力
type UpdateCommentInput struct {
	Content string `json:"content" binding:"required"`
}
