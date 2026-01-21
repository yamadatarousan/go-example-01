package repository

import (
	"context"
	"database/sql"
	"fmt"

	// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
	// 例: "gin-quickstart/backend/internal/domain" → "gin-quickstart/backend/internal/domain"
	"gin-quickstart/backend/internal/domain"

	"github.com/lib/pq"
)

// tagRepository はTagRepositoryインターフェースの実装
type tagRepository struct {
	db *sql.DB
}

// NewTagRepository はtagRepositoryの新しいインスタンスを作成
func NewTagRepository(db *sql.DB) TagRepository {
	return &tagRepository{
		db: db,
	}
}

// FindOrCreateByName はタグを名前で検索し、存在しなければ作成
func (r *tagRepository) FindOrCreateByName(ctx context.Context, name string) (domain.Tag, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, created_at
	`

	var tag domain.Tag
	err := r.db.QueryRowContext(ctx, query, name).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("タグの作成/取得に失敗しました: %w", err)
	}

	return tag, nil
}

// FindAll は全タグを取得
func (r *tagRepository) FindAll(ctx context.Context) ([]domain.Tag, error) {
	query := `
		SELECT id, name, created_at
		FROM tags
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("タグ一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var tag domain.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("タグのスキャンに失敗しました: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("タグ一覧の取得に失敗しました: %w", err)
	}

	return tags, nil
}

// AttachToTodo はTODOにタグを紐付け
func (r *tagRepository) AttachToTodo(ctx context.Context, todoID int, tagIDs []int) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// 既存の紐付けを削除
	if err := r.DetachFromTodo(ctx, todoID); err != nil {
		return err
	}

	// 新しい紐付けを一括挿入
	query := `
		INSERT INTO todo_tags (todo_id, tag_id)
		SELECT $1, unnest($2::int[])
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, todoID, pq.Array(tagIDs))
	if err != nil {
		return fmt.Errorf("タグの紐付けに失敗しました: %w", err)
	}

	return nil
}

// DetachFromTodo はTODOからすべてのタグを解除
func (r *tagRepository) DetachFromTodo(ctx context.Context, todoID int) error {
	query := `DELETE FROM todo_tags WHERE todo_id = $1`

	_, err := r.db.ExecContext(ctx, query, todoID)
	if err != nil {
		return fmt.Errorf("タグの解除に失敗しました: %w", err)
	}

	return nil
}

// FindByTodoID はTODO IDに紐づくタグを取得
func (r *tagRepository) FindByTodoID(ctx context.Context, todoID int) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.name, t.created_at
		FROM tags t
		INNER JOIN todo_tags tt ON t.id = tt.tag_id
		WHERE tt.todo_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, todoID)
	if err != nil {
		return nil, fmt.Errorf("タグの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var tag domain.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("タグのスキャンに失敗しました: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("タグの取得に失敗しました: %w", err)
	}

	return tags, nil
}
