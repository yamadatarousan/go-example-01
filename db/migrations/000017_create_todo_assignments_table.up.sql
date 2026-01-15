CREATE TABLE IF NOT EXISTS todo_assignments (
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (todo_id, user_id)
);

-- インデックス
CREATE INDEX idx_todo_assignments_todo ON todo_assignments(todo_id);
CREATE INDEX idx_todo_assignments_user ON todo_assignments(user_id);
