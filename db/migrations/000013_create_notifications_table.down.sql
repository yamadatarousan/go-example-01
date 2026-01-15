-- Phase 4: 通知テーブルの削除

DROP INDEX IF EXISTS idx_notifications_todo;
DROP INDEX IF EXISTS idx_notifications_type;
DROP INDEX IF EXISTS idx_notifications_user;
DROP TABLE IF EXISTS notifications;
