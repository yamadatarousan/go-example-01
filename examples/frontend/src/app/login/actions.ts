// ログイン用 Server Actions
// サーバーサイドで実行され、API との通信と Cookie の設定を行う

'use server'

import { cookies } from 'next/headers'
import { fetchWithoutAuth } from '@/lib/server-api'

// ログインリクエストの型
interface LoginRequest {
  email: string
  password: string
}

// ログインレスポンスの型
interface LoginResponse {
  token: string
  user: {
    id: number
    email: string
    role: string
  }
}

// Server Action の戻り値の型
interface ActionResult {
  error?: string
  success?: boolean
}

/**
 * ログイン処理を行う Server Action
 * バックエンド API を呼び出し、JWT トークンを httpOnly Cookie に保存する
 */
export async function loginAction(data: LoginRequest): Promise<ActionResult> {
  const result = await fetchWithoutAuth<LoginResponse>('/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })

  if (result.error) {
    return { error: result.error }
  }

  if (!result.data?.token) {
    return { error: 'トークンが取得できませんでした' }
  }

  // JWT トークンを httpOnly Cookie に保存
  // セキュリティのため、JavaScript からアクセスできないようにする
  const cookieStore = await cookies()
  cookieStore.set('token', result.data.token, {
    httpOnly: true,        // JavaScript からアクセス不可
    secure: process.env.NODE_ENV === 'production', // HTTPS のみ（本番環境）
    sameSite: 'lax',       // CSRF 対策
    maxAge: 60 * 60 * 24,  // 24 時間
    path: '/',             // 全パスで有効
  })

  return { success: true }
}
