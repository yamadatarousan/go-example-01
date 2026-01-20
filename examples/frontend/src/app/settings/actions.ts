// 設定画面用 Server Actions

'use server'

import { cookies } from 'next/headers'
import { fetchWithAuth, mockData, isMockMode } from '@/lib/server-api'

// ユーザープロフィールの型定義
export interface UserProfile {
  id: number
  username: string
  email: string
  bio?: string
  image?: string | null
  role: 'user' | 'admin'
  created_at: string
}

// プロフィール更新リクエスト
interface UpdateProfileRequest {
  username?: string
  bio?: string
}

/**
 * ユーザープロフィールを取得する Server Action
 * 注意: バックエンドに /api/v1/users/me がないため、現在はモックデータを返す
 */
export async function getProfileAction(): Promise<{
  data?: UserProfile
  error?: string
}> {
  // バックエンドにエンドポイントがないため、モックデータを返す
  // TODO: バックエンド実装後に fetchWithAuth('/api/v1/users/me') に変更
  return { data: mockData.userProfile }
}

/**
 * ユーザープロフィールを更新する Server Action
 * 注意: バックエンドに PUT /api/v1/users/me がないため、現在はモック処理
 */
export async function updateProfileAction(data: UpdateProfileRequest): Promise<{
  data?: UserProfile
  error?: string
}> {
  // バックエンドにエンドポイントがないため、モック処理
  // TODO: バックエンド実装後に fetchWithAuth('/api/v1/users/me', { method: 'PUT', ... }) に変更

  // 仮の成功レスポンス
  return {
    data: {
      ...mockData.userProfile,
      username: data.username || mockData.userProfile.username,
      bio: data.bio || mockData.userProfile.bio,
    },
  }
}

/**
 * ログアウト Server Action
 * Cookie から JWT トークンを削除する
 */
export async function logoutAction(): Promise<void> {
  const cookieStore = await cookies()
  cookieStore.delete('token')
}
