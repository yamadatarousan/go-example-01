package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gin-quickstart/examples/backend/internal/domain"
)

// todoRepository はTodoRepositoryインターフェースの実装
type todoRepository struct {
	db *sql.DB
}

// NewTodoRepository はTodoRepositoryの新しいインスタンスを作成
func NewTodoRepository(db *sql.DB) TodoRepository {
	return &todoRepository{db: db}
}

// FindAll は指定されたユーザーの全てのTODOを取得（Phase 2拡張フィールド含む）
func (r *todoRepository) FindAll(ctx context.Context, userID int) ([]domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		err := rows.Scan(&todo.ID, &todo.Name, &todo.Description, &todo.Status, &todo.Priority, &todo.DueDate, &todo.UserID, &todo.CategoryID, &todo.ParentTodoID, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// FindByID は指定されたIDのTODOを取得（Phase 2拡張フィールド含む）
// userID=0の場合はシステム権限としてuser_idチェックをスキップ（バックグラウンドワーカー用）
func (r *todoRepository) FindByID(ctx context.Context, todoID, userID int) (domain.Todo, error) {
	var todo domain.Todo
	var query string
	var args []interface{}

	if userID == 0 {
		// システム権限: user_idチェックなし
		query = `SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos WHERE id = $1`
		args = []interface{}{todoID}
	} else {
		// 通常のユーザー権限: user_idチェックあり
		query = `SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos WHERE id = $1 AND user_id = $2`
		args = []interface{}{todoID, userID}
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&todo.ID, &todo.Name, &todo.Description, &todo.Status, &todo.Priority,
		&todo.DueDate, &todo.UserID, &todo.CategoryID, &todo.ParentTodoID,
		&todo.CreatedAt, &todo.UpdatedAt,
	)
	if err != nil {
		return todo, err
	}
	return todo, nil
}

// Create はTODOを作成（Phase 2拡張フィールド含む）
func (r *todoRepository) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	var id int
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO todos (name, user_id, priority, status, description, due_date, category_id, parent_todo_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		todo.Name, todo.UserID, todo.Priority, todo.Status, todo.Description, todo.DueDate, todo.CategoryID, todo.ParentTodoID,
	).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id
	return todo, nil
}

// execTx はトランザクションを実行するためのヘルパー関数
// トランザクションを開始し、渡された関数(fn)を実行します。
// fnがエラーを返した場合、トランザクションはロールバックされます。
// エラーがなければ、トランザクションはコミットされます。
func (r *todoRepository) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		// エラーが発生した場合、ロールバックを試みる
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// createTodoInTx はトランザクション内でTODOと監査ログを作成
func (r *todoRepository) createTodoInTx(tx *sql.Tx, todo domain.Todo) (domain.Todo, error) {
	// 1. todosテーブルに新しいTODOを挿入し、IDを取得（Phase 2拡張フィールド含む）
	var id int
	err := tx.QueryRow(`
		INSERT INTO todos (name, user_id, priority, status, description, due_date, category_id, parent_todo_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, todo.Name, todo.UserID, todo.Priority, todo.Status, todo.Description, todo.DueDate, todo.CategoryID, todo.ParentTodoID).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id

	// 2. todo_audit_logsテーブルに監査ログを挿入
	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", id, "create")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

// CreateTodoWithAudit はトランザクションを使用してTODOと監査ログを作成
func (r *todoRepository) CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	var createdTodo domain.Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		createdTodo, err = r.createTodoInTx(tx, todo)
		return err
	})

	return createdTodo, err
}

// updateTodoInTx はトランザクション内でTODOを更新し、監査ログを作成
func (r *todoRepository) updateTodoInTx(tx *sql.Tx, todo domain.Todo) (domain.Todo, error) {
	// 1. todosテーブルのレコードを更新
	result, err := tx.Exec(
		"UPDATE todos SET name = $1 WHERE id = $2 AND user_id = $3",
		todo.Name, todo.ID, todo.UserID,
	)
	if err != nil {
		return todo, err
	}

	// 更新された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return todo, err
	}
	if rowsAffected == 0 {
		return todo, sql.ErrNoRows
	}

	// 2. todo_audit_logsテーブルに監査ログを挿入
	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", todo.ID, "update")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

// UpdateTodoWithAudit はトランザクションを使用してTODOを更新
func (r *todoRepository) UpdateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	var updatedTodo domain.Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		updatedTodo, err = r.updateTodoInTx(tx, todo)
		return err
	})

	return updatedTodo, err
}

// deleteTodoInTx はトランザクション内でTODOを削除し、監査ログを作成
func (r *todoRepository) deleteTodoInTx(tx *sql.Tx, todoID, userID int) error {
	// 1. 監査ログを先に挿入（TODOが削除される前に）
	_, err := tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", todoID, "delete")
	if err != nil {
		return err
	}

	// 2. todosテーブルからレコードを削除
	result, err := tx.Exec("DELETE FROM todos WHERE id = $1 AND user_id = $2", todoID, userID)
	if err != nil {
		return err
	}

	// 削除された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteTodoWithAudit はトランザクションを使用してTODOを削除
func (r *todoRepository) DeleteTodoWithAudit(ctx context.Context, todoID, userID int) error {
	return r.execTx(ctx, func(tx *sql.Tx) error {
		return r.deleteTodoInTx(tx, todoID, userID)
	})
}

// ============================================================================
// Phase 2で追加されたメソッド
// ============================================================================

// UpdateStatus はTODOのステータスを更新
func (r *todoRepository) UpdateStatus(ctx context.Context, todoID, userID int, status string) error {
	query := `UPDATE todos SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`

	result, err := r.db.ExecContext(ctx, query, status, todoID, userID)
	if err != nil {
		return fmt.Errorf("ステータスの更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// FindOverdue は期限切れのTODOを取得
func (r *todoRepository) FindOverdue(ctx context.Context, userID int) ([]domain.Todo, error) {
	query := `
		SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		  AND due_date < NOW()
		  AND status != 'done'
		ORDER BY due_date ASC
	`

	return r.queryTodos(ctx, query, userID)
}

// FindToday は今日が期限のTODOを取得
func (r *todoRepository) FindToday(ctx context.Context, userID int) ([]domain.Todo, error) {
	query := `
		SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		  AND DATE(due_date) = CURRENT_DATE
		  AND status != 'done'
		ORDER BY priority DESC, due_date ASC
	`

	return r.queryTodos(ctx, query, userID)
}

// FindThisWeek は今週が期限のTODOを取得
func (r *todoRepository) FindThisWeek(ctx context.Context, userID int) ([]domain.Todo, error) {
	query := `
		SELECT id, name, description, status, priority, due_date, user_id, category_id, parent_todo_id, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		  AND due_date >= CURRENT_DATE
		  AND due_date < CURRENT_DATE + INTERVAL '7 days'
		  AND status != 'done'
		ORDER BY due_date ASC, priority DESC
	`

	return r.queryTodos(ctx, query, userID)
}

// queryTodos は共通のクエリ実行ロジック
func (r *todoRepository) queryTodos(ctx context.Context, query string, args ...interface{}) ([]domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("TODOの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var todos []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Name,
			&todo.Description,
			&todo.Status,
			&todo.Priority,
			&todo.DueDate,
			&todo.UserID,
			&todo.CategoryID,
			&todo.ParentTodoID,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("TODOのスキャンに失敗しました: %w", err)
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("TODOの取得に失敗しました: %w", err)
	}

	return todos, nil
}

// ============================================================================
// Phase 3で追加されたメソッド
// ============================================================================

// Search は高度な検索・フィルタリング機能
func (r *todoRepository) Search(ctx context.Context, userID int, filters domain.SearchFilters) (domain.SearchResult, error) {
	// デフォルト値の設定
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 10
	}
	if filters.Sort == "" {
		filters.Sort = "created_at"
	}
	if filters.Order == "" {
		filters.Order = "desc"
	}

	// ベースクエリ
	baseQuery := `
		FROM todos t
		WHERE t.user_id = $1
	`

	// WHERE条件を動的に構築
	conditions := []string{"t.user_id = $1"}
	args := []interface{}{userID}
	argCount := 1

	// ステータスフィルター
	if filters.Status != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argCount))
		args = append(args, *filters.Status)
	}

	// 優先度フィルター
	if filters.Priority != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argCount))
		args = append(args, *filters.Priority)
	}

	// カテゴリーフィルター
	if filters.CategoryID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.category_id = $%d", argCount))
		args = append(args, *filters.CategoryID)
	}

	// タグフィルター
	if len(filters.TagIDs) > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf(`
			t.id IN (
				SELECT todo_id FROM todo_tags WHERE tag_id = ANY($%d)
			)
		`, argCount))
		args = append(args, filters.TagIDs)
	}

	// 全文検索
	if filters.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.search_vector @@ plainto_tsquery('english', $%d)", argCount))
		args = append(args, filters.Search)
	}

	// 期限範囲フィルター
	if filters.DueFrom != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.due_date >= $%d", argCount))
		args = append(args, *filters.DueFrom)
	}
	if filters.DueTo != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("t.due_date <= $%d", argCount))
		args = append(args, *filters.DueTo)
	}

	// WHERE句を構築
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// 総件数を取得
	countQuery := "SELECT COUNT(*) " + baseQuery
	if len(conditions) > 1 {
		countQuery = "SELECT COUNT(*) FROM todos t " + whereClause
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("総件数の取得に失敗しました: %w", err)
	}

	// ソート順の検証
	allowedSorts := map[string]bool{
		"due_date":   true,
		"priority":   true,
		"created_at": true,
		"updated_at": true,
	}
	if !allowedSorts[filters.Sort] {
		filters.Sort = "created_at"
	}
	if filters.Order != "asc" && filters.Order != "desc" {
		filters.Order = "desc"
	}

	// データ取得クエリ
	offset := (filters.Page - 1) * filters.Limit
	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.name, t.description, t.status, t.priority, t.due_date,
		       t.user_id, t.category_id, t.parent_todo_id, t.created_at, t.updated_at
		FROM todos t
		%s
		ORDER BY t.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, filters.Sort, filters.Order, argCount+1, argCount+2)

	args = append(args, filters.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("TODOの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var todos []domain.Todo
	for rows.Next() {
		var todo domain.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Name,
			&todo.Description,
			&todo.Status,
			&todo.Priority,
			&todo.DueDate,
			&todo.UserID,
			&todo.CategoryID,
			&todo.ParentTodoID,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return domain.SearchResult{}, fmt.Errorf("TODOのスキャンに失敗しました: %w", err)
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return domain.SearchResult{}, fmt.Errorf("TODOの取得に失敗しました: %w", err)
	}

	// 総ページ数を計算
	totalPages := (total + filters.Limit - 1) / filters.Limit

	return domain.SearchResult{
		Todos:      todos,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetStatistics はTODO統計情報を取得
func (r *todoRepository) GetStatistics(ctx context.Context, userID int) (domain.Statistics, error) {
	var stats domain.Statistics

	// 総件数とステータス別カウント
	statusQuery := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'todo' THEN 1 END) as todo_count,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_count,
			COUNT(CASE WHEN status = 'done' THEN 1 END) as done_count
		FROM todos
		WHERE user_id = $1
	`

	var todoCount, inProgressCount, doneCount int
	err := r.db.QueryRowContext(ctx, statusQuery, userID).Scan(
		&stats.TotalCount,
		&todoCount,
		&inProgressCount,
		&doneCount,
	)
	if err != nil {
		return stats, fmt.Errorf("ステータス別カウントの取得に失敗しました: %w", err)
	}

	stats.StatusCounts = map[string]int{
		"todo":        todoCount,
		"in_progress": inProgressCount,
		"done":        doneCount,
	}

	// 優先度別カウント
	priorityQuery := `
		SELECT
			COUNT(CASE WHEN priority = 'low' THEN 1 END) as low_count,
			COUNT(CASE WHEN priority = 'medium' THEN 1 END) as medium_count,
			COUNT(CASE WHEN priority = 'high' THEN 1 END) as high_count
		FROM todos
		WHERE user_id = $1
	`

	var lowCount, mediumCount, highCount int
	err = r.db.QueryRowContext(ctx, priorityQuery, userID).Scan(
		&lowCount,
		&mediumCount,
		&highCount,
	)
	if err != nil {
		return stats, fmt.Errorf("優先度別カウントの取得に失敗しました: %w", err)
	}

	stats.PriorityCounts = map[string]int{
		"low":    lowCount,
		"medium": mediumCount,
		"high":   highCount,
	}

	// 期限切れカウント
	overdueQuery := `
		SELECT COUNT(*)
		FROM todos
		WHERE user_id = $1
		  AND due_date < NOW()
		  AND status != 'done'
	`
	err = r.db.QueryRowContext(ctx, overdueQuery, userID).Scan(&stats.OverdueCount)
	if err != nil {
		return stats, fmt.Errorf("期限切れカウントの取得に失敗しました: %w", err)
	}

	// 今日期限カウント
	todayQuery := `
		SELECT COUNT(*)
		FROM todos
		WHERE user_id = $1
		  AND DATE(due_date) = CURRENT_DATE
		  AND status != 'done'
	`
	err = r.db.QueryRowContext(ctx, todayQuery, userID).Scan(&stats.DueTodayCount)
	if err != nil {
		return stats, fmt.Errorf("今日期限カウントの取得に失敗しました: %w", err)
	}

	// 今週期限カウント
	weekQuery := `
		SELECT COUNT(*)
		FROM todos
		WHERE user_id = $1
		  AND due_date >= CURRENT_DATE
		  AND due_date < CURRENT_DATE + INTERVAL '7 days'
		  AND status != 'done'
	`
	err = r.db.QueryRowContext(ctx, weekQuery, userID).Scan(&stats.DueThisWeekCount)
	if err != nil {
		return stats, fmt.Errorf("今週期限カウントの取得に失敗しました: %w", err)
	}

	return stats, nil
}
