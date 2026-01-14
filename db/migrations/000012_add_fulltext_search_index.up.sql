-- Phase 3: 全文検索用のインデックスを追加

-- tsvector列を追加（名前と説明を結合した全文検索用）
ALTER TABLE todos ADD COLUMN search_vector tsvector;

-- tsvectorを自動更新するトリガー関数を作成
CREATE OR REPLACE FUNCTION todos_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- INSERT/UPDATEで自動的にsearch_vectorを更新するトリガーを作成
CREATE TRIGGER todos_search_vector_trigger
BEFORE INSERT OR UPDATE ON todos
FOR EACH ROW
EXECUTE FUNCTION todos_search_vector_update();

-- 既存のレコードのsearch_vectorを更新
UPDATE todos SET search_vector =
  setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
  setweight(to_tsvector('english', COALESCE(description, '')), 'B');

-- 全文検索用のGINインデックスを作成
CREATE INDEX idx_todos_search_vector ON todos USING GIN(search_vector);
