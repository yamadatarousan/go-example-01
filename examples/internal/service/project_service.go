// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/examples/internal/domain" → "gin-quickstart/internal/domain"

package service

import (
	"context"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
)

// ProjectService はプロジェクト関連のビジネスロジックを提供
type ProjectService struct {
	projectRepo repository.ProjectRepository
}

// NewProjectService はProjectServiceの新しいインスタンスを作成
func NewProjectService(projectRepo repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
	}
}

// CreateProject はプロジェクトを作成
func (s *ProjectService) CreateProject(ctx context.Context, input domain.CreateProjectInput, ownerID int) (domain.Project, error) {
	return s.projectRepo.Create(ctx, input, ownerID)
}

// GetProjects は指定されたユーザーが参加している全てのプロジェクトを取得
func (s *ProjectService) GetProjects(ctx context.Context, userID int) ([]domain.Project, error) {
	return s.projectRepo.FindAll(ctx, userID)
}

// GetProject は指定されたIDのプロジェクトを取得
func (s *ProjectService) GetProject(ctx context.Context, projectID, userID int) (domain.Project, error) {
	return s.projectRepo.FindByID(ctx, projectID, userID)
}

// UpdateProject はプロジェクト情報を更新
func (s *ProjectService) UpdateProject(ctx context.Context, projectID int, input domain.UpdateProjectInput, userID int) (domain.Project, error) {
	return s.projectRepo.Update(ctx, projectID, input, userID)
}

// DeleteProject はプロジェクトを削除
func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID int) error {
	return s.projectRepo.Delete(ctx, projectID, userID)
}

// AddMember はプロジェクトにメンバーを追加
func (s *ProjectService) AddMember(ctx context.Context, projectID int, input domain.AddMemberInput, requesterID int) error {
	return s.projectRepo.AddMember(ctx, projectID, input, requesterID)
}

// RemoveMember はプロジェクトからメンバーを削除
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID, requesterID int) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID, requesterID)
}

// GetMembers はプロジェクトの全メンバーを取得
func (s *ProjectService) GetMembers(ctx context.Context, projectID, userID int) ([]domain.ProjectMember, error) {
	return s.projectRepo.GetMembers(ctx, projectID, userID)
}

// UpdateMemberRole はメンバーの役割を更新
func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, targetUserID int, newRole string, requesterID int) error {
	return s.projectRepo.UpdateMemberRole(ctx, projectID, targetUserID, newRole, requesterID)
}
