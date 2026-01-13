-- カテゴリーテーブルの作成

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),  -- カラーコード（例: #FF5733）
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ユーザーごとのカテゴリー名はユニーク
CREATE UNIQUE INDEX idx_categories_user_name ON categories(user_id, name);

-- TODOテーブルにカテゴリー外部キーを追加
ALTER TABLE todos ADD COLUMN category_id INT REFERENCES categories(id) ON DELETE SET NULL;

-- インデックスの追加
CREATE INDEX idx_todos_category ON todos(category_id) WHERE category_id IS NOT NULL;
