-- TODO機能拡張のロールバック

-- インデックスの削除
DROP INDEX IF EXISTS idx_todos_parent;
DROP INDEX IF EXISTS idx_todos_due_date;
DROP INDEX IF EXISTS idx_todos_priority;
DROP INDEX IF EXISTS idx_todos_status;

-- カラムの削除
ALTER TABLE todos DROP COLUMN IF EXISTS parent_todo_id;
ALTER TABLE todos DROP COLUMN IF EXISTS description;
ALTER TABLE todos DROP COLUMN IF EXISTS status;
ALTER TABLE todos DROP COLUMN IF EXISTS due_date;
ALTER TABLE todos DROP COLUMN IF EXISTS priority;
