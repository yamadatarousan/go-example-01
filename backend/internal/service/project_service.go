// 例: "gin-quickstart/backend/internal/domain" → "gin-quickstart/backend/internal/domain"

package service

import (
	"context"
	"errors"

	"gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/repository"
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
	// プロジェクトを取得
	project, err := s.projectRepo.FindByID(ctx, projectID, userID)
	if err != nil {
		return domain.Project{}, err
	}

	// ドメインロジックで権限チェック
	if !project.CanBeUpdatedBy(userID) {
		return domain.Project{}, errors.New("only project owner can update project")
	}

	return s.projectRepo.Update(ctx, projectID, input)
}

// DeleteProject はプロジェクトを削除
func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID int) error {
	// プロジェクトを取得
	project, err := s.projectRepo.FindByID(ctx, projectID, userID)
	if err != nil {
		return err
	}

	// ドメインロジックで権限チェック
	if !project.CanBeDeletedBy(userID) {
		return errors.New("only project owner can delete project")
	}

	return s.projectRepo.Delete(ctx, projectID)
}

// AddMember はプロジェクトにメンバーを追加
func (s *ProjectService) AddMember(ctx context.Context, projectID int, input domain.AddMemberInput, requesterID int) error {
	// リクエスターがオーナーまたは管理者であることを確認
	role, err := s.projectRepo.GetRole(ctx, projectID, requesterID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return errors.New("only owner or admin can add members")
	}

	return s.projectRepo.AddMember(ctx, projectID, input)
}

// RemoveMember はプロジェクトからメンバーを削除
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID, requesterID int) error {
	// リクエスターがオーナーまたは管理者であることを確認
	role, err := s.projectRepo.GetRole(ctx, projectID, requesterID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return errors.New("only owner or admin can remove members")
	}

	// オーナーは削除できない
	targetRole, err := s.projectRepo.GetRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		return errors.New("cannot remove project owner")
	}

	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

// GetMembers はプロジェクトの全メンバーを取得
func (s *ProjectService) GetMembers(ctx context.Context, projectID, userID int) ([]domain.ProjectMember, error) {
	// メンバーであることを確認
	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("access denied: not a member of this project")
	}

	return s.projectRepo.GetMembers(ctx, projectID)
}

// UpdateMemberRole はメンバーの役割を更新
func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, targetUserID int, newRole string, requesterID int) error {
	// リクエスターがオーナーであることを確認
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, requesterID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only project owner can update member roles")
	}

	// オーナーの役割は変更できない
	targetRole, err := s.projectRepo.GetRole(ctx, projectID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		return errors.New("cannot change owner role")
	}

	return s.projectRepo.UpdateMemberRole(ctx, projectID, targetUserID, newRole)
}
