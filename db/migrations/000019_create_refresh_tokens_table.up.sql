-- リフレッシュトークンテーブル
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE
);

-- ユーザーIDでのクエリを高速化
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- 有効期限切れトークンの削除クエリを高速化
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
