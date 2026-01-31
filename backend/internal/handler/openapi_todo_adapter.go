package handler

import (
	openapi "gin-quickstart/backend/internal/openapi/gen"

	"github.com/gin-gonic/gin"
)

// TodoOpenAPIAdapter bridges generated OpenAPI handlers to existing TodoHandler.
type TodoOpenAPIAdapter struct {
	todoHandler *TodoHandler
}

// NewTodoOpenAPIAdapter creates a new adapter for TodoHandler.
func NewTodoOpenAPIAdapter(todoHandler *TodoHandler) *TodoOpenAPIAdapter {
	return &TodoOpenAPIAdapter{
		todoHandler: todoHandler,
	}
}

// GetTodos handles GET /api/v1/todos.
func (a *TodoOpenAPIAdapter) GetTodos(c *gin.Context) {
	ErrorHandler(a.todoHandler.GetTodos)(c)
}

// CreateTodo handles POST /api/v1/todos.
func (a *TodoOpenAPIAdapter) CreateTodo(c *gin.Context) {
	ErrorHandler(a.todoHandler.CreateTodo)(c)
}

// GetOverdueTodos handles GET /api/v1/todos/overdue.
func (a *TodoOpenAPIAdapter) GetOverdueTodos(c *gin.Context) {
	ErrorHandler(a.todoHandler.GetOverdueTodos)(c)
}

// SearchTodos handles GET /api/v1/todos/search.
func (a *TodoOpenAPIAdapter) SearchTodos(c *gin.Context, _ openapi.SearchTodosParams) {
	ErrorHandler(a.todoHandler.SearchTodos)(c)
}

// GetTodoStatistics handles GET /api/v1/todos/statistics.
func (a *TodoOpenAPIAdapter) GetTodoStatistics(c *gin.Context) {
	ErrorHandler(a.todoHandler.GetStatistics)(c)
}

// GetTodayTodos handles GET /api/v1/todos/today.
func (a *TodoOpenAPIAdapter) GetTodayTodos(c *gin.Context) {
	ErrorHandler(a.todoHandler.GetTodayTodos)(c)
}

// GetThisWeekTodos handles GET /api/v1/todos/week.
func (a *TodoOpenAPIAdapter) GetThisWeekTodos(c *gin.Context) {
	ErrorHandler(a.todoHandler.GetThisWeekTodos)(c)
}

// DeleteTodo handles DELETE /api/v1/todos/{id}.
func (a *TodoOpenAPIAdapter) DeleteTodo(c *gin.Context, _ openapi.TodoId) {
	ErrorHandler(a.todoHandler.DeleteTodo)(c)
}

// GetTodo handles GET /api/v1/todos/{id}.
func (a *TodoOpenAPIAdapter) GetTodo(c *gin.Context, _ openapi.TodoId) {
	ErrorHandler(a.todoHandler.GetTodo)(c)
}

// UpdateTodo handles PUT /api/v1/todos/{id}.
func (a *TodoOpenAPIAdapter) UpdateTodo(c *gin.Context, _ openapi.TodoId) {
	ErrorHandler(a.todoHandler.UpdateTodo)(c)
}

// CompleteTodo handles POST /api/v1/todos/{id}/complete.
func (a *TodoOpenAPIAdapter) CompleteTodo(c *gin.Context, _ openapi.TodoId) {
	ErrorHandler(a.todoHandler.CompleteTodo)(c)
}

// ReopenTodo handles POST /api/v1/todos/{id}/reopen.
func (a *TodoOpenAPIAdapter) ReopenTodo(c *gin.Context, _ openapi.TodoId) {
	ErrorHandler(a.todoHandler.ReopenTodo)(c)
}
