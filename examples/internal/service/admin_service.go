package service

import (
	"context"

	"gin-quickstart/examples/internal/domain"
	"gin-quickstart/examples/internal/repository"
)

// AdminService は管理者関連のビジネスロジックを提供
type AdminService struct {
	userRepo repository.UserRepository
}

// NewAdminService はAdminServiceの新しいインスタンスを作成
func NewAdminService(userRepo repository.UserRepository) *AdminService {
	return &AdminService{
		userRepo: userRepo,
	}
}

// GetAllUsers は全てのユーザーを取得
func (s *AdminService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.FindAllUsers(ctx)
}
