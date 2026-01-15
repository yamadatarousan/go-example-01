-- Phase 4: 通知テーブルの作成

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    todo_id INT REFERENCES todos(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- 'deadline_reminder', 'todo_assigned', 'todo_completed'
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 通知取得の高速化
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);

-- 通知タイプでのフィルタリング用
CREATE INDEX idx_notifications_type ON notifications(type);

-- TODOに紐づく通知の取得用
CREATE INDEX idx_notifications_todo ON notifications(todo_id) WHERE todo_id IS NOT NULL;
