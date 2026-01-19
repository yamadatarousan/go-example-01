-- パフォーマンス改善のためのインデックス（Phase 6）

-- TODOsテーブル: ユーザーIDとステータスの複合インデックス（よく一緒に検索される）
CREATE INDEX IF NOT EXISTS idx_todos_user_status ON todos(user_id, status);

-- TODOsテーブル: 期限日のインデックス（期限切れTODOの検索を高速化）
CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;

-- TODOsテーブル: プロジェクトIDのインデックス（プロジェクト別TODO検索を高速化）
CREATE INDEX IF NOT EXISTS idx_todos_project_id ON todos(project_id) WHERE project_id IS NOT NULL;

-- Categoriesテーブル: ユーザーIDのインデックス（ユーザー別カテゴリー検索を高速化）
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);

-- Notificationsテーブル: ユーザーIDと既読状態の複合インデックス
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);

-- Remindersテーブル: リマインド時刻と送信状態の複合インデックス（バックグラウンドワーカー用）
CREATE INDEX IF NOT EXISTS idx_reminders_remind_at_sent ON reminders(remind_at, is_sent) WHERE is_sent = FALSE;

-- Project Membersテーブル: ユーザーIDのインデックス（ユーザーが参加しているプロジェクト検索を高速化）
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);

-- Commentsテーブル: TODOIDのインデックス（TODO別コメント検索を高速化）
CREATE INDEX IF NOT EXISTS idx_comments_todo_id ON comments(todo_id);

-- Todo Audit Logsテーブル: TODOIDと作成日時の複合インデックス（監査ログ検索を高速化）
CREATE INDEX IF NOT EXISTS idx_todo_audit_logs_todo_created ON todo_audit_logs(todo_id, created_at);
