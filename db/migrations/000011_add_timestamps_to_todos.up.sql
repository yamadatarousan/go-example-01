-- TODOsテーブルにタイムスタンプカラムを追加

ALTER TABLE todos ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE todos ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- インデックスの追加
CREATE INDEX idx_todos_created_at ON todos(created_at);
