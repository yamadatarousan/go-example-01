// API レスポンスとリクエストの型定義
// バックエンドの Go 構造体と対応させる

// =============================================================================
// ユーザー関連の型
// =============================================================================

/** ユーザー情報 */
export interface User {
  id: number
  email: string
  created_at: string
  updated_at: string
}

/** ログインリクエスト */
export interface LoginRequest {
  email: string
  password: string
}

/** サインアップリクエスト */
export interface SignupRequest {
  email: string
  password: string
}

/** 認証レスポンス */
export interface AuthResponse {
  token: string
  user: User
}

// =============================================================================
// TODO 関連の型
// =============================================================================

/** TODO アイテム */
export interface Todo {
  id: number
  user_id: number
  title: string
  completed: boolean
  created_at: string
  updated_at: string
}

/** TODO 作成リクエスト */
export interface CreateTodoRequest {
  title: string
}

/** TODO 更新リクエスト */
export interface UpdateTodoRequest {
  title?: string
  completed?: boolean
}

/** TODO 一覧レスポンス */
export interface TodosResponse {
  todos: Todo[]
}

// =============================================================================
// API エラーレスポンス
// =============================================================================

/** API エラー */
export interface ApiError {
  error: string
}

// =============================================================================
// Server Action 共通型
// =============================================================================

/**
 * Server Action の統一された戻り値型
 * 成功時は data を、失敗時は error を返す
 */
export type ActionResult<T = void> =
  | { success: true; data?: T }
  | { success: false; error: string }

// =============================================================================
// SSE イベント型（リアルタイム通知用）
// =============================================================================

/** SSE イベントの種類 */
export type SSEEventType = 'todo_created' | 'todo_updated' | 'todo_deleted'

/** SSE イベントペイロード */
export interface SSEEvent {
  type: SSEEventType
  data: Todo | { id: number }
}
