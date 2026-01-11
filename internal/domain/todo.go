package domain

// Todo represents a todo item in the domain model
type Todo struct {
	ID     int    `json:"id"`
	Name   string `json:"name" binding:"required"`
	UserID int    `json:"user_id"`
}
