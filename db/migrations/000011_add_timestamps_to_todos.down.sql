-- TODOsテーブルからタイムスタンプカラムを削除

DROP INDEX IF EXISTS idx_todos_created_at;
ALTER TABLE todos DROP COLUMN IF EXISTS updated_at;
ALTER TABLE todos DROP COLUMN IF EXISTS created_at;
