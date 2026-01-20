// ログイン用 Server Actions
// サーバーサイドで実行され、API との通信と Cookie の設定を行う

'use server'

import { cookies } from 'next/headers'

// API のベース URL
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

// ログインリクエストの型
interface LoginRequest {
  email: string
  password: string
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
  try {
    // バックエンド API にログインリクエストを送信
    const response = await fetch(`${API_BASE_URL}/api/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })

    // エラーレスポンスの処理
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return {
        error: errorData.error || 'ログインに失敗しました',
      }
    }

    // 成功時: レスポンスから JWT トークンを取得
    const result = await response.json()

    // JWT トークンを httpOnly Cookie に保存
    // セキュリティのため、JavaScript からアクセスできないようにする
    const cookieStore = await cookies()
    cookieStore.set('token', result.token, {
      httpOnly: true,        // JavaScript からアクセス不可
      secure: process.env.NODE_ENV === 'production', // HTTPS のみ（本番環境）
      sameSite: 'lax',       // CSRF 対策
      maxAge: 60 * 60 * 24,  // 24 時間
      path: '/',             // 全パスで有効
    })

    return { success: true }
  } catch (error) {
    console.error('Login error:', error)
    return {
      error: 'サーバーとの通信に失敗しました',
    }
  }
}
