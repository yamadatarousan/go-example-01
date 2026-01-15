DROP INDEX IF EXISTS idx_todos_project;
DROP INDEX IF EXISTS idx_projects_owner;

ALTER TABLE todos DROP COLUMN IF EXISTS project_id;

DROP TABLE IF EXISTS projects;
