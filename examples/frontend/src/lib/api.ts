// API クライアント
// バックエンドとの通信を行うユーティリティ関数

import type {
  LoginRequest,
  SignupRequest,
  AuthResponse,
  Todo,
  TodosResponse,
  CreateTodoRequest,
  UpdateTodoRequest,
  ApiError,
} from '@/types'

// APIのベースURL（環境変数から取得、デフォルトは localhost:8080）
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

// =============================================================================
// 共通のfetch ラッパー
// =============================================================================

/**
 * API リクエストを送信する共通関数
 * エラーハンドリングと認証ヘッダーの付与を行う
 */
async function fetchApi<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    // Cookie を含める（認証用）
    credentials: 'include',
  })

  // レスポンスが OK でない場合はエラーをスロー
  if (!response.ok) {
    const errorData: ApiError = await response.json().catch(() => ({
      error: `HTTP error! status: ${response.status}`,
    }))
    throw new Error(errorData.error)
  }

  // 204 No Content の場合は空オブジェクトを返す
  if (response.status === 204) {
    return {} as T
  }

  return response.json()
}

// =============================================================================
// 認証 API
// =============================================================================

/** ログイン */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  return fetchApi<AuthResponse>('/api/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

/** サインアップ */
export async function signup(data: SignupRequest): Promise<AuthResponse> {
  return fetchApi<AuthResponse>('/api/signup', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

/** ログアウト */
export async function logout(): Promise<void> {
  return fetchApi<void>('/api/logout', {
    method: 'POST',
  })
}

// =============================================================================
// TODO API
// =============================================================================

/** TODO 一覧取得 */
export async function getTodos(): Promise<Todo[]> {
  const response = await fetchApi<TodosResponse>('/api/todos', {
    method: 'GET',
  })
  return response.todos
}

/** TODO 作成 */
export async function createTodo(data: CreateTodoRequest): Promise<Todo> {
  return fetchApi<Todo>('/api/todos', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

/** TODO 更新 */
export async function updateTodo(
  id: number,
  data: UpdateTodoRequest
): Promise<Todo> {
  return fetchApi<Todo>(`/api/todos/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

/** TODO 削除 */
export async function deleteTodo(id: number): Promise<void> {
  return fetchApi<void>(`/api/todos/${id}`, {
    method: 'DELETE',
  })
}

// =============================================================================
// SSE（Server-Sent Events）接続
// =============================================================================

/**
 * SSE 接続を確立する
 * リアルタイム通知を受信するための EventSource を返す
 */
export function connectSSE(): EventSource {
  const url = `${API_BASE_URL}/api/events`
  return new EventSource(url, { withCredentials: true })
}
