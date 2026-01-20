// 型定義
// バックエンドのドメインモデルに対応するTypeScript型

// ============================================================================
// ユーザー関連
// ============================================================================

export type User = {
  id: number;
  email: string;
  role: string;
  created_at: string;
};

export type LoginInput = {
  email: string;
  password: string;
};

export type SignupInput = {
  email: string;
  password: string;
};

// ============================================================================
// TODO関連
// ============================================================================

export type TodoStatus = "todo" | "in_progress" | "done";
export type TodoPriority = "low" | "medium" | "high";

export type Todo = {
  id: number;
  name: string;
  description?: string | null;
  status: TodoStatus;
  priority: TodoPriority;
  due_date?: string | null;
  user_id: number;
  category_id?: number | null;
  parent_todo_id?: number | null;
  project_id?: number | null;
  created_at: string;
  updated_at: string;
  category?: Category | null;
  tags?: Tag[];
};

export type CreateTodoInput = {
  name: string;
  description?: string | null;
  status?: TodoStatus;
  priority?: TodoPriority;
  due_date?: string | null;
  category_id?: number | null;
  parent_todo_id?: number | null;
  project_id?: number | null;
  tag_ids?: number[];
};

export type UpdateTodoInput = {
  name?: string;
  description?: string | null;
  status?: TodoStatus;
  priority?: TodoPriority;
  due_date?: string | null;
  category_id?: number | null;
  parent_todo_id?: number | null;
  project_id?: number | null;
  tag_ids?: number[];
};

// ============================================================================
// カテゴリー関連
// ============================================================================

export type Category = {
  id: number;
  name: string;
  color?: string | null;
  user_id: number;
  created_at: string;
  updated_at: string;
};

// ============================================================================
// タグ関連
// ============================================================================

export type Tag = {
  id: number;
  name: string;
};

// ============================================================================
// プロジェクト関連
// ============================================================================

export type Project = {
  id: number;
  name: string;
  description?: string | null;
  owner_id: number;
  created_at: string;
  updated_at: string;
};

export type CreateProjectInput = {
  name: string;
  description?: string | null;
};

// ============================================================================
// 検索・統計関連
// ============================================================================

export type SearchResult<T> = {
  todos: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
};

export type Statistics = {
  total_count: number;
  status_counts: Record<string, number>;
  priority_counts: Record<string, number>;
  overdue_count: number;
  due_today_count: number;
  due_this_week_count: number;
};

// ============================================================================
// API共通
// ============================================================================

export type ApiError = {
  error: string;
  message?: string;
};

export type ActionResult<T> = {
  success: true;
  data: T;
} | {
  success: false;
  error: string;
};
