-- タグ機能の実装（Many-to-Many関係）

-- タグテーブルの作成
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- TODOとタグの中間テーブル（Many-to-Many）
CREATE TABLE IF NOT EXISTS todo_tags (
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (todo_id, tag_id)
);

-- インデックスの追加（検索パフォーマンス向上）
CREATE INDEX idx_todo_tags_todo ON todo_tags(todo_id);
CREATE INDEX idx_todo_tags_tag ON todo_tags(tag_id);
CREATE INDEX idx_tags_name ON tags(name);
