// Server Actions 用 API クライアント
// サーバーサイドで実行され、Cookie から JWT を取得して Authorization ヘッダーに付与する

import { cookies } from 'next/headers'

// API のベース URL
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

// モックモードフラグ（バックエンド未実装のエンドポイント用）
const USE_MOCK = false

// =============================================================================
// 共通の認証付き fetch 関数
// =============================================================================

/**
 * Cookie から JWT トークンを取得して Authorization ヘッダーを生成する
 */
async function getAuthHeaders(): Promise<HeadersInit> {
  const cookieStore = await cookies()
  const token = cookieStore.get('token')?.value

  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  }
}

/**
 * 認証付き API リクエストを送信する共通関数
 * Server Actions から使用する
 */
export async function fetchWithAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<{ data?: T; error?: string }> {
  try {
    const headers = await getAuthHeaders()
    const url = `${API_BASE_URL}${endpoint}`

    const response = await fetch(url, {
      ...options,
      headers: {
        ...headers,
        ...options.headers,
      },
    })

    // エラーレスポンスの処理
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return { error: errorData.error || `HTTP error! status: ${response.status}` }
    }

    // 204 No Content の場合は空オブジェクトを返す
    if (response.status === 204) {
      return { data: {} as T }
    }

    const data = await response.json()
    return { data }
  } catch (error) {
    console.error('API request error:', error)
    return { error: 'サーバーとの通信に失敗しました' }
  }
}

/**
 * 認証不要の API リクエストを送信する共通関数
 * ログイン・サインアップ用
 */
export async function fetchWithoutAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<{ data?: T; error?: string }> {
  try {
    const url = `${API_BASE_URL}${endpoint}`

    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    })

    // エラーレスポンスの処理
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return { error: errorData.error || `HTTP error! status: ${response.status}` }
    }

    const data = await response.json()
    return { data }
  } catch (error) {
    console.error('API request error:', error)
    return { error: 'サーバーとの通信に失敗しました' }
  }
}

// =============================================================================
// モックデータ（バックエンド未実装のエンドポイント用）
// =============================================================================

export const mockData = {
  // ユーザープロフィール（GET /api/v1/users/me 用）
  userProfile: {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    bio: 'これはテストユーザーです',
    image: null,
    role: 'user' as const,
    created_at: '2024-01-01T00:00:00Z',
  },

  // 統計情報（GET /api/v1/todos/statistics 用）
  statistics: {
    total: 10,
    by_status: {
      todo: 5,
      in_progress: 3,
      done: 2,
    },
    by_priority: {
      high: 2,
      medium: 5,
      low: 3,
    },
    overdue: 1,
    due_today: 2,
    due_this_week: 4,
  },
}

/**
 * モックモードかどうかを確認する
 */
export function isMockMode(): boolean {
  return USE_MOCK
}
