package service

import (
	"context"
	"fmt"

	// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
	// 例: "gin-quickstart/backend/internal/domain" → "gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/repository"
)

// CategoryService はカテゴリー関連のビジネスロジックを提供
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService はCategoryServiceの新しいインスタンスを作成
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory はカテゴリーを作成
func (s *CategoryService) CreateCategory(ctx context.Context, userID int, input domain.CreateCategoryInput) (domain.Category, error) {
	category := domain.Category{
		Name:   input.Name,
		Color:  input.Color,
		UserID: userID,
	}

	created, err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの作成に失敗しました: %w", err)
	}

	return created, nil
}

// GetCategories はユーザーの全カテゴリーを取得
func (s *CategoryService) GetCategories(ctx context.Context, userID int) ([]domain.Category, error) {
	categories, err := s.categoryRepo.FindAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("カテゴリー一覧の取得に失敗しました: %w", err)
	}

	return categories, nil
}

// GetCategory は特定のカテゴリーを取得
func (s *CategoryService) GetCategory(ctx context.Context, categoryID, userID int) (domain.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, categoryID, userID)
	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの取得に失敗しました: %w", err)
	}

	return category, nil
}

// UpdateCategory はカテゴリーを更新
func (s *CategoryService) UpdateCategory(ctx context.Context, categoryID, userID int, input domain.UpdateCategoryInput) (domain.Category, error) {
	updated, err := s.categoryRepo.Update(ctx, categoryID, userID, input)
	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの更新に失敗しました: %w", err)
	}

	return updated, nil
}

// DeleteCategory はカテゴリーを削除
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID, userID int) error {
	err := s.categoryRepo.Delete(ctx, categoryID, userID)
	if err != nil {
		return fmt.Errorf("カテゴリーの削除に失敗しました: %w", err)
	}

	return nil
}
