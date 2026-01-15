// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package domain

import "time"

// Project はプロジェクトを表すドメインモデル
type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     int       `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectInput はプロジェクト作成時の入力
type CreateProjectInput struct {
	Name        string  `json:"name" binding:"required,max=200"`
	Description *string `json:"description"`
}

// UpdateProjectInput はプロジェクト更新時の入力
type UpdateProjectInput struct {
	Name        *string `json:"name" binding:"omitempty,max=200"`
	Description *string `json:"description"`
}
