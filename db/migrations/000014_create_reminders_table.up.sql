-- Phase 4: リマインダーテーブルの作成

CREATE TABLE IF NOT EXISTS reminders (
    id SERIAL PRIMARY KEY,
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    remind_at TIMESTAMPTZ NOT NULL,
    is_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- リマインダー送信対象の検索用（未送信かつ送信時刻が過去）
CREATE INDEX idx_reminders_pending ON reminders(is_sent, remind_at) WHERE is_sent = FALSE;

-- TODOに紐づくリマインダー取得用
CREATE INDEX idx_reminders_todo ON reminders(todo_id);
