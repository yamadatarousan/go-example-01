package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/service"

	"github.com/gin-gonic/gin"
)

// TodoHandler はTODO関連のHTTPリクエストを処理
type TodoHandler struct {
  todoService *service.TodoService
}

// NewTodoHandler はTodoHandlerの新しいインスタンスを作成
func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
  return &TodoHandler{
    todoService: todoService,
  }
}

// GetTodos は指定されたユーザーの全てのTODOを取得
func (h *TodoHandler) GetTodos(c *gin.Context) error {
  claims := c.MustGet("claims").(*service.AppClaims)
  userID, _ := strconv.Atoi(claims.Subject)

  todos, err := h.todoService.GetTodos(userID)
  if err != nil {
    return err
  }

  c.JSON(http.StatusOK, todos)
  return nil
}

// GetTodo は指定されたIDのTODOを取得
func (h *TodoHandler) GetTodo(c *gin.Context) error {
  // URLパラメータからTODO IDを取得
  todoID, err := strconv.Atoi(c.Param("id"))
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
    return nil
  }

  // ログインユーザーのIDを取得
  claims := c.MustGet("claims").(*service.AppClaims)
  userID, _ := strconv.Atoi(claims.Subject)

  // TODOを取得
  todo, err := h.todoService.GetTodo(todoID, userID)
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
      return nil
    }
    return err
  }

  c.JSON(http.StatusOK, todo)
  return nil
}

// CreateTodo はTODOを作成
func (h *TodoHandler) CreateTodo(c *gin.Context) error {
  var newTodo domain.Todo
  if err := c.ShouldBindJSON(&newTodo); err != nil {
    return err
  }

  claims := c.MustGet("claims").(*service.AppClaims)
  userID, _ := strconv.Atoi(claims.Subject)
  newTodo.UserID = userID

  createdTodo, err := h.todoService.CreateTodo(c.Request.Context(), newTodo)
  if err != nil {
    return err
  }

  c.JSON(http.StatusCreated, createdTodo)
  return nil
}

// UpdateTodo はTODOを更新
func (h *TodoHandler) UpdateTodo(c *gin.Context) error {
  // URLパラメータからTODO IDを取得
  todoID, err := strconv.Atoi(c.Param("id"))
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
    return nil
  }

  // リクエストボディから更新内容を取得
  var updateTodo domain.Todo
  if err := c.BindJSON(&updateTodo); err != nil {
    return err
  }

  // ログインユーザーのIDを取得
  claims := c.MustGet("claims").(*service.AppClaims)
  userID, _ := strconv.Atoi(claims.Subject)

  // 更新対象のTODOを設定
  updateTodo.ID = todoID
  updateTodo.UserID = userID

  // TODOを更新
  updatedTodo, err := h.todoService.UpdateTodo(c.Request.Context(), updateTodo)
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
      return nil
    }
    return nil
  }

  c.JSON(http.StatusOK, updatedTodo)
  return nil
}

// DeleteTodo はTODOを削除
func (h *TodoHandler) DeleteTodo(c *gin.Context) error {
▸-// URLパラメータからTODO IDを取得
▸-todoID, err := strconv.Atoi(c.Param("id"))
▸-if err != nil {
▸-▸-c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
▸-▸-return nil
▸-}

▸-// ログインユーザーのIDを取得
▸-claims := c.MustGet("claims").(*service.AppClaims)
▸-userID, _ := strconv.Atoi(claims.Subject)

▸-// TODOを削除
▸-err = h.todoService.DeleteTodo(c.Request.Context(), todoID, userID)
▸-if err != nil {
▸-▸-if errors.Is(err, sql.ErrNoRows) {
▸-▸-▸-c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
▸-▸-▸-return nil
▸-▸-}
▸-▸-return err
▸-}

▸-c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
▸-return nil
}
