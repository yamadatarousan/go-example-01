// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、import pathから "examples/" を削除してください ★★★
// 例: "gin-quickstart/internal/domain" → "gin-quickstart/internal/domain"

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gin-quickstart/internal/domain"
)

// projectRepository はProjectRepositoryインターフェースの実装
type projectRepository struct {
	db *sql.DB
}

// NewProjectRepository はProjectRepositoryの新しいインスタンスを作成
func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create は新しいプロジェクトを作成
func (r *projectRepository) Create(ctx context.Context, input domain.CreateProjectInput, ownerID int) (domain.Project, error) {
	var project domain.Project

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return project, err
	}
	defer tx.Rollback()

	// プロジェクトを作成
	err = tx.QueryRowContext(ctx, `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, owner_id, created_at, updated_at
	`, input.Name, input.Description, ownerID).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return project, err
	}

	// オーナーをメンバーとして追加
	_, err = tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, project.ID, ownerID)
	if err != nil {
		return project, err
	}

	if err = tx.Commit(); err != nil {
		return project, err
	}

	return project, nil
}

// FindAll は指定されたユーザーが参加している全てのプロジェクトを取得
func (r *projectRepository) FindAll(ctx context.Context, userID int) ([]domain.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var project domain.Project
		err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// FindByID は指定されたIDのプロジェクトを取得（メンバーチェック付き）
func (r *projectRepository) FindByID(ctx context.Context, projectID, userID int) (domain.Project, error) {
	var project domain.Project

	// メンバーであることを確認しながら取得
	err := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE p.id = $1 AND pm.user_id = $2
	`, projectID, userID).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return project, errors.New("project not found or access denied")
		}
		return project, err
	}

	return project, nil
}

// Update はプロジェクト情報を更新
func (r *projectRepository) Update(ctx context.Context, projectID int, input domain.UpdateProjectInput) (domain.Project, error) {
	var project domain.Project

	// 更新クエリを動的に構築
	query := "UPDATE projects SET updated_at = NOW()"
	args := []interface{}{}
	argPosition := 1

	if input.Name != nil {
		argPosition++
		query += ", name = $" + fmt.Sprint(argPosition)
		args = append(args, *input.Name)
	}
	if input.Description != nil {
		argPosition++
		query += ", description = $" + fmt.Sprint(argPosition)
		args = append(args, *input.Description)
	}

	query += " WHERE id = $1 RETURNING id, name, description, owner_id, created_at, updated_at"
	args = append([]interface{}{projectID}, args...)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return project, err
	}

	return project, nil
}

// Delete はプロジェクトを削除
func (r *projectRepository) Delete(ctx context.Context, projectID int) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM projects WHERE id = $1
	`, projectID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("project not found")
	}

	return nil
}

// AddMember はプロジェクトにメンバーを追加
func (r *projectRepository) AddMember(ctx context.Context, projectID int, input domain.AddMemberInput) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
	`, projectID, input.UserID, input.Role)
	return err
}

// RemoveMember はプロジェクトからメンバーを削除
func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID int) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("member not found")
	}

	return nil
}

// GetMembers はプロジェクトの全メンバーを取得
func (r *projectRepository) GetMembers(ctx context.Context, projectID int) ([]domain.ProjectMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT project_id, user_id, role, joined_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY joined_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.ProjectMember
	for rows.Next() {
		var member domain.ProjectMember
		err := rows.Scan(&member.ProjectID, &member.UserID, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// UpdateMemberRole はメンバーの役割を更新
func (r *projectRepository) UpdateMemberRole(ctx context.Context, projectID, targetUserID int, newRole string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE project_members
		SET role = $1
		WHERE project_id = $2 AND user_id = $3
	`, newRole, projectID, targetUserID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("member not found")
	}

	return nil
}

// IsOwner はユーザーがプロジェクトのオーナーかどうかを確認
func (r *projectRepository) IsOwner(ctx context.Context, projectID, userID int) (bool, error) {
	var ownerID int
	err := r.db.QueryRowContext(ctx, `
		SELECT owner_id FROM projects WHERE id = $1
	`, projectID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("project not found")
		}
		return false, err
	}

	return ownerID == userID, nil
}

// IsMember はユーザーがプロジェクトのメンバーかどうかを確認
func (r *projectRepository) IsMember(ctx context.Context, projectID, userID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetRole はユーザーのプロジェクト内での役割を取得
func (r *projectRepository) GetRole(ctx context.Context, projectID, userID int) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("user is not a member of this project")
		}
		return "", err
	}
	return role, nil
}
