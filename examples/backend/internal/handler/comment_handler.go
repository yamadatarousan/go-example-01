// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/backend/internal/domain" → "gin-quickstart/examples/backend/internal/domain"

package handler

import (
	"net/http"
	"strconv"

	"gin-quickstart/examples/backend/internal/domain"
	"gin-quickstart/examples/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentHandler はコメント関連のHTTPリクエストを処理
type CommentHandler struct {
	commentService *service.CommentService
}

// NewCommentHandler はCommentHandlerの新しいインスタンスを作成
func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment はコメントを作成
func (h *CommentHandler) CreateComment(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	var input domain.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	comment, err := h.commentService.CreateComment(c.Request.Context(), todoID, input, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, comment)
	return nil
}

// GetCommentsByTodoID は指定されたTODOの全てのコメントを取得
func (h *CommentHandler) GetCommentsByTodoID(c *gin.Context) error {
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	comments, err := h.commentService.GetCommentsByTodoID(c.Request.Context(), todoID, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, comments)
	return nil
}

// GetComment は指定されたIDのコメントを取得
func (h *CommentHandler) GetComment(c *gin.Context) error {
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	comment, err := h.commentService.GetComment(c.Request.Context(), commentID, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, comment)
	return nil
}

// UpdateComment はコメントを更新
func (h *CommentHandler) UpdateComment(c *gin.Context) error {
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	var input domain.UpdateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), commentID, input, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, comment)
	return nil
}

// DeleteComment はコメントを削除
func (h *CommentHandler) DeleteComment(c *gin.Context) error {
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	err = h.commentService.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}
