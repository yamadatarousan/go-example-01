package handler

import (
	"fmt"
	"net/http"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler はユーザー関連のHTTPリクエストを処理
type UserHandler struct {
	authService *service.AuthService
}

// NewUserHandler はUserHandlerの新しいインスタンスを作成
func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

// Signup はユーザー登録を処理
func (h *UserHandler) Signup(c *gin.Context) error {
	var input domain.SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	createdUser, err := h.authService.Signup(c.Request.Context(), input)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         createdUser.ID,
		"email":      createdUser.Email,
		"created_at": createdUser.CreatedAt,
	})
	return nil
}

// Login はユーザーログインを処理
func (h *UserHandler) Login(c *gin.Context) error {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	// アクセストークンの生成
	token, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		return err
	}

	// トークンからユーザーIDを取得
	claims, err := h.authService.ParseToken(token)
	if err != nil {
		return err
	}

	// リフレッシュトークンの生成
	userID := 0
	if _, err := fmt.Sscanf(claims.Subject, "%d", &userID); err != nil || userID == 0 {
		return fmt.Errorf("invalid user ID in token: %w", err)
	}
	refreshToken, err := h.authService.GenerateRefreshToken(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
	})
	return nil
}

// RefreshToken はリフレッシュトークンを使って新しいアクセストークンを取得
func (h *UserHandler) RefreshToken(c *gin.Context) error {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	// 新しいアクセストークンを生成
	newAccessToken, err := h.authService.RefreshAccessToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
	return nil
}

// RevokeRefreshToken はリフレッシュトークンを無効化（ログアウト）
func (h *UserHandler) RevokeRefreshToken(c *gin.Context) error {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	err := h.authService.RevokeRefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Refresh token revoked successfully",
	})
	return nil
}
