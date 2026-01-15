// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package domain

import "time"

// ProjectMember はプロジェクトメンバーを表すドメインモデル
type ProjectMember struct {
	ProjectID int       `json:"project_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"` // 'owner', 'admin', 'member'
	JoinedAt  time.Time `json:"joined_at"`
}

// AddMemberInput はメンバー追加時の入力
type AddMemberInput struct {
	UserID int    `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner admin member"`
}
