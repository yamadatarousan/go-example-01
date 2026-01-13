package handler

import (
	"net/http"
	"strconv"

	// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
	// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"
	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler はカテゴリー関連のHTTPリクエストを処理
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler はCategoryHandlerの新しいインスタンスを作成
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategory はカテゴリーを作成
func (h *CategoryHandler) CreateCategory(c *gin.Context) error {
	var input domain.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	// JWTから取得したユーザーIDを使用
	userID, exists := c.Get("user_id")
	if !exists {
		return service.ErrUnauthorized
	}

	created, err := h.categoryService.CreateCategory(c.Request.Context(), userID.(int), input)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, created)
	return nil
}

// GetCategories はユーザーの全カテゴリーを取得
func (h *CategoryHandler) GetCategories(c *gin.Context) error {
	userID, exists := c.Get("user_id")
	if !exists {
		return service.ErrUnauthorized
	}

	categories, err := h.categoryService.GetCategories(userID.(int))
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, categories)
	return nil
}

// GetCategory は特定のカテゴリーを取得
func (h *CategoryHandler) GetCategory(c *gin.Context) error {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return service.ErrUnauthorized
	}

	category, err := h.categoryService.GetCategory(categoryID, userID.(int))
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, category)
	return nil
}

// UpdateCategory はカテゴリーを更新
func (h *CategoryHandler) UpdateCategory(c *gin.Context) error {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	var input domain.UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return service.ErrUnauthorized
	}

	updated, err := h.categoryService.UpdateCategory(c.Request.Context(), categoryID, userID.(int), input)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, updated)
	return nil
}

// DeleteCategory はカテゴリーを削除
func (h *CategoryHandler) DeleteCategory(c *gin.Context) error {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return service.ErrUnauthorized
	}

	err = h.categoryService.DeleteCategory(c.Request.Context(), categoryID, userID.(int))
	if err != nil {
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}
