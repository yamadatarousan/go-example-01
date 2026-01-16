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

// ============================================================================
// ドメインロジック（アクセス制御）
// ============================================================================

// IsOwnedBy は指定されたユーザーがプロジェクトのオーナーかどうかを判定
func (p *Project) IsOwnedBy(userID int) bool {
	return p.OwnerID == userID
}

// CanBeUpdatedBy は指定されたユーザーがプロジェクトを更新できるかを判定
//
// ビジネスルール:
// - オーナーのみ更新可能
func (p *Project) CanBeUpdatedBy(userID int) bool {
	return p.IsOwnedBy(userID)
}

// CanBeDeletedBy は指定されたユーザーがプロジェクトを削除できるかを判定
//
// ビジネスルール:
// - オーナーのみ削除可能
func (p *Project) CanBeDeletedBy(userID int) bool {
	return p.IsOwnedBy(userID)
}
