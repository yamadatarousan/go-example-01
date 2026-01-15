CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- TODOテーブルにproject_idカラムを追加
ALTER TABLE todos ADD COLUMN IF NOT EXISTS project_id INT REFERENCES projects(id) ON DELETE SET NULL;

-- インデックス
CREATE INDEX idx_projects_owner ON projects(owner_id);
CREATE INDEX idx_todos_project ON todos(project_id) WHERE project_id IS NOT NULL;
