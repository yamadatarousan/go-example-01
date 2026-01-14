-- Phase 3: 全文検索用のインデックスを削除

-- GINインデックスを削除
DROP INDEX IF EXISTS idx_todos_search_vector;

-- トリガーを削除
DROP TRIGGER IF EXISTS todos_search_vector_trigger ON todos;

-- トリガー関数を削除
DROP FUNCTION IF EXISTS todos_search_vector_update();

-- tsvector列を削除
ALTER TABLE todos DROP COLUMN IF EXISTS search_vector;
