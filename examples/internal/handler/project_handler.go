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

// ProjectHandler はプロジェクト関連のHTTPリクエストを処理
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler はProjectHandlerの新しいインスタンスを作成
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject はプロジェクトを作成
func (h *ProjectHandler) CreateProject(c *gin.Context) error {
	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	var input domain.CreateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), input, userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, project)
	return nil
}

// GetProjects は指定されたユーザーが参加している全てのプロジェクトを取得
func (h *ProjectHandler) GetProjects(c *gin.Context) error {
	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	projects, err := h.projectService.GetProjects(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, projects)
	return nil
}

// GetProject は指定されたIDのプロジェクトを取得
func (h *ProjectHandler) GetProject(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	project, err := h.projectService.GetProject(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, project)
	return nil
}

// UpdateProject はプロジェクト情報を更新
func (h *ProjectHandler) UpdateProject(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	var input domain.UpdateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), projectID, input, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, project)
	return nil
}

// DeleteProject はプロジェクトを削除
func (h *ProjectHandler) DeleteProject(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	err = h.projectService.DeleteProject(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

// AddMember はプロジェクトにメンバーを追加
func (h *ProjectHandler) AddMember(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	var input domain.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	err = h.projectService.AddMember(c.Request.Context(), projectID, input, requesterID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
	return nil
}

// RemoveMember はプロジェクトからメンバーを削除
func (h *ProjectHandler) RemoveMember(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	err = h.projectService.RemoveMember(c.Request.Context(), projectID, userID, requesterID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

// GetMembers はプロジェクトの全メンバーを取得
func (h *ProjectHandler) GetMembers(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)

	members, err := h.projectService.GetMembers(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, members)
	return nil
}

// UpdateMemberRole はメンバーの役割を更新
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) error {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	targetUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return domain.ErrInvalidInput

		return nil
	}

	claims := c.MustGet("claims").(*service.AppClaims)
	requesterID, _ := strconv.Atoi(claims.Subject)

	var input struct {
		Role string `json:"role" binding:"required,oneof=owner admin member"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	err = h.projectService.UpdateMemberRole(c.Request.Context(), projectID, targetUserID, input.Role, requesterID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member role updated successfully"})
	return nil
}
