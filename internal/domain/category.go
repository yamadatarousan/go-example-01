package domain

import "time"

// Category はTODOのカテゴリー
type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Color     *string   `json:"color" binding:"omitempty,hexcolor"` // カラーコード（例: #FF5733）
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryInput はカテゴリー作成時の入力
type CreateCategoryInput struct {
	Name  string  `json:"name" binding:"required"`
	Color *string `json:"color" binding:"omitempty,hexcolor"`
}

// UpdateCategoryInput はカテゴリー更新時の入力
type UpdateCategoryInput struct {
	Name  *string `json:"name"`
	Color *string `json:"color" binding:"omitempty,hexcolor"`
}
