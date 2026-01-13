-- TODO機能の拡張: 優先度、期限、ステータス、説明、サブタスク機能を追加

-- 優先度フィールド (high/medium/low)
ALTER TABLE todos ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';

-- 期限フィールド
ALTER TABLE todos ADD COLUMN due_date TIMESTAMPTZ;

-- ステータスフィールド (todo/in_progress/done)
ALTER TABLE todos ADD COLUMN status VARCHAR(20) DEFAULT 'todo';

-- 詳細説明フィールド
ALTER TABLE todos ADD COLUMN description TEXT;

-- サブタスク機能（自己参照）
ALTER TABLE todos ADD COLUMN parent_todo_id INT REFERENCES todos(id) ON DELETE CASCADE;

-- インデックスの追加（パフォーマンス最適化）
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_todos_parent ON todos(parent_todo_id) WHERE parent_todo_id IS NOT NULL;
