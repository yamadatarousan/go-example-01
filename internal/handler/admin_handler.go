package handler

import (
  "net/http"

  "gin-quickstart/internal/service"

  "github.com/gin-gonic/gin"
)

// AdminHandler は管理者関連のHTTPリクエストを処理
type AdminHandler struct {
  adminService *service.AdminService
}

// NewAdminHandler はAdminHandlerの新しいインスタンスを作成
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
  return &AdminHandler {
    adminService: adminService,
  }
}

// GetAllUsers は全てのユーザーを取得
func (h *AdminHandler) GetAllUsers(c *gin.Context) error {
  users, err := h.adminService.GetAllUsers()
  if err != nil {
    return err
  }
  
  c.JSON(http.StatusOK, users)
  return nil
}
