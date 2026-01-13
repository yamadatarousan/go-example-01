-- タグ機能のロールバック

-- インデックスの削除
DROP INDEX IF EXISTS idx_tags_name;
DROP INDEX IF EXISTS idx_todo_tags_tag;
DROP INDEX IF EXISTS idx_todo_tags_todo;

-- テーブルの削除
DROP TABLE IF EXISTS todo_tags;
DROP TABLE IF EXISTS tags;
