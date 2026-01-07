package main

import (
	"database/sql"
)

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) FindAll(userID int) ([]Todo, error) {
	rows, err := r.db.Query("SELECT id, name, user_id FROM todos WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Name, &todo.UserID)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (r *TodoRepository) Create(todo Todo) (Todo, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO todos (name, user_id) VALUES ($1, $2) RETURNING id",
		todo.Name, todo.UserID,
	).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id
	return todo, nil
}
