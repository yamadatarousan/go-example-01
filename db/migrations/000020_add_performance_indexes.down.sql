-- パフォーマンスインデックスの削除
DROP INDEX IF EXISTS idx_todo_audit_logs_todo_created;
DROP INDEX IF EXISTS idx_comments_todo_id;
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_reminders_remind_at_sent;
DROP INDEX IF EXISTS idx_notifications_user_read;
DROP INDEX IF EXISTS idx_categories_user_id;
DROP INDEX IF EXISTS idx_todos_project_id;
DROP INDEX IF EXISTS idx_todos_due_date;
DROP INDEX IF EXISTS idx_todos_user_status;
