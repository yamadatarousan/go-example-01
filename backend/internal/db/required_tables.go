package db

// RequiredTables は起動時に存在を確認するテーブル一覧
var RequiredTables = []string{
	"users",
	"todos",
	"todo_audit_logs",
	"categories",
	"tags",
	"todo_tags",
	"notifications",
	"reminders",
	"projects",
	"project_members",
	"todo_assignments",
	"comments",
	"refresh_tokens",
}
