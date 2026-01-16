// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/examples/internal/domain"

package handler

import (
	"net/http"
	"strconv"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
)

// TodoAssignmentHandler はTODO担当者関連のHTTPリクエストを処理
type TodoAssignmentHandler struct {
	assignmentService *service.TodoAssignmentService
}

// NewTodoAssignmentHandler はTodoAssignmentHandlerの新しいインスタンスを作成
func NewTodoAssignmentHandler(assignmentService *service.TodoAssignmentService) *TodoAssignmentHandler {
	return &TodoAssignmentHandler{
		assignmentService: assignmentService,
	}
}

// AssignUser はTODOにユーザーを担当者として割り当て
func (h *TodoAssignmentHandler) AssignUser(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	var input domain.AssignUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	assignment, err := h.assignmentService.AssignUser(c.Request.Context(), todoID, input, requesterID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, assignment)
	return nil
}

// UnassignUser はTODOから担当者を解除
func (h *TodoAssignmentHandler) UnassignUser(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	err = h.assignmentService.UnassignUser(c.Request.Context(), todoID, userID, requesterID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

// GetAssignments は指定されたTODOの全ての担当者を取得
func (h *TodoAssignmentHandler) GetAssignments(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	assignments, err := h.assignmentService.GetAssignments(c.Request.Context(), todoID, requesterID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, assignments)
	return nil
}
