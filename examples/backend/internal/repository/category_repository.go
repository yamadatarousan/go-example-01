package repository

import (
	"context"
	"database/sql"
	"fmt"

	// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
	// 例: "gin-quickstart/examples/backend/internal/domain" → "gin-quickstart/backend/internal/domain"
	"gin-quickstart/examples/backend/internal/domain"
)

// categoryRepository はCategoryRepositoryインターフェースの実装
type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository はcategoryRepositoryの新しいインスタンスを作成
func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

// Create はカテゴリーを新規作成
func (r *categoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	query := `
		INSERT INTO categories (name, color, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var createdCategory domain.Category
	createdCategory.Name = category.Name
	createdCategory.Color = category.Color
	createdCategory.UserID = category.UserID

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Color,
		category.UserID,
	).Scan(&createdCategory.ID, &createdCategory.CreatedAt, &createdCategory.UpdatedAt)

	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの作成に失敗しました: %w", err)
	}

	return createdCategory, nil
}

// FindAll はユーザーの全カテゴリーを取得
func (r *categoryRepository) FindAll(ctx context.Context, userID int) ([]domain.Category, error) {
	query := `
		SELECT id, name, color, user_id, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("カテゴリー一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.UserID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("カテゴリーのスキャンに失敗しました: %w", err)
		}
		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("カテゴリー一覧の取得に失敗しました: %w", err)
	}

	return categories, nil
}

// FindByID は特定のカテゴリーを取得
func (r *categoryRepository) FindByID(ctx context.Context, categoryID, userID int) (domain.Category, error) {
	query := `
		SELECT id, name, color, user_id, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, categoryID, userID).Scan(
		&category.ID,
		&category.Name,
		&category.Color,
		&category.UserID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return domain.Category{}, ErrNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの取得に失敗しました: %w", err)
	}

	return category, nil
}

// Update はカテゴリーを更新
func (r *categoryRepository) Update(ctx context.Context, categoryID, userID int, input domain.UpdateCategoryInput) (domain.Category, error) {
	// 既存のカテゴリーを取得
	existing, err := r.FindByID(ctx, categoryID, userID)
	if err != nil {
		return domain.Category{}, err
	}

	// 更新するフィールドを決定
	name := existing.Name
	if input.Name != nil {
		name = *input.Name
	}

	color := existing.Color
	if input.Color != nil {
		color = input.Color
	}

	query := `
		UPDATE categories
		SET name = $1, color = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, name, color, user_id, created_at, updated_at
	`

	var updated domain.Category
	err = r.db.QueryRowContext(ctx, query, name, color, categoryID, userID).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Color,
		&updated.UserID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		return domain.Category{}, fmt.Errorf("カテゴリーの更新に失敗しました: %w", err)
	}

	return updated, nil
}

// Delete はカテゴリーを削除
func (r *categoryRepository) Delete(ctx context.Context, categoryID, userID int) error {
	// カテゴリーが存在するか確認
	_, err := r.FindByID(ctx, categoryID, userID)
	if err != nil {
		return err
	}

	query := `DELETE FROM categories WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, categoryID, userID)
	if err != nil {
		return fmt.Errorf("カテゴリーの削除に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
