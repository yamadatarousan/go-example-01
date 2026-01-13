package domain

import "time"

// Tag はTODOに付与するタグ
type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagInput はタグ作成時の入力
type CreateTagInput struct {
	Name string `json:"name" binding:"required"`
}
