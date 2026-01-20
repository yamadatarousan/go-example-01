// サインアップ用 Server Actions
// サーバーサイドで実行され、API との通信と Cookie の設定を行う

'use server'

import { cookies } from 'next/headers'

// API のベース URL
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

// サインアップリクエストの型
interface SignupRequest {
  email: string
  password: string
}

// Server Action の戻り値の型
interface ActionResult {
  error?: string
  success?: boolean
}

/**
 * サインアップ処理を行う Server Action
 * バックエンド API を呼び出し、JWT トークンを httpOnly Cookie に保存する
 */
export async function signupAction(data: SignupRequest): Promise<ActionResult> {
  try {
    // バックエンド API にサインアップリクエストを送信
    const response = await fetch(`${API_BASE_URL}/api/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })

    // エラーレスポンスの処理
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))

      // よくあるエラーメッセージの日本語化
      if (errorData.error?.includes('already exists')) {
        return { error: 'このメールアドレスは既に登録されています' }
      }

      return {
        error: errorData.error || 'サインアップに失敗しました',
      }
    }

    // 成功時: レスポンスから JWT トークンを取得
    const result = await response.json()

    // JWT トークンを httpOnly Cookie に保存
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
    console.error('Signup error:', error)
    return {
      error: 'サーバーとの通信に失敗しました',
    }
  }
}
