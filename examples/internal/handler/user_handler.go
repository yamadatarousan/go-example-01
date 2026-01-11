package handler

import (
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

	createdUser, err := h.authService.Signup(input)
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

	token, err := h.authService.Login(input)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}
