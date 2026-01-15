-- Phase 4: リマインダーテーブルの削除

DROP INDEX IF EXISTS idx_reminders_todo;
DROP INDEX IF EXISTS idx_reminders_pending;
DROP TABLE IF EXISTS reminders;
