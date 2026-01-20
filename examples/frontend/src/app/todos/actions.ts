// TODO 関連の Server Actions
// サーバーサイドで実行され、API との通信を行う

'use server'

import { cookies } from 'next/headers'
import type { Todo } from '@/types'

// API のベース URL
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

// 共通の型定義
interface TodoActionResult {
  error?: string
  todo?: Todo
}

interface DeleteActionResult {
  error?: string
  success?: boolean
}

/**
 * 認証ヘッダーを取得するヘルパー関数
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
 * TODO を作成する Server Action
 */
export async function createTodoAction(title: string): Promise<TodoActionResult> {
  try {
    const headers = await getAuthHeaders()

    const response = await fetch(`${API_BASE_URL}/api/todos`, {
      method: 'POST',
      headers,
      body: JSON.stringify({ title }),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return { error: errorData.error || 'TODO の作成に失敗しました' }
    }

    const todo = await response.json()
    return { todo }
  } catch (error) {
    console.error('Create todo error:', error)
    return { error: 'サーバーとの通信に失敗しました' }
  }
}

/**
 * TODO を更新する Server Action
 */
export async function updateTodoAction(
  id: number,
  updates: { title?: string; completed?: boolean }
): Promise<TodoActionResult> {
  try {
    const headers = await getAuthHeaders()

    const response = await fetch(`${API_BASE_URL}/api/todos/${id}`, {
      method: 'PUT',
      headers,
      body: JSON.stringify(updates),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return { error: errorData.error || 'TODO の更新に失敗しました' }
    }

    const todo = await response.json()
    return { todo }
  } catch (error) {
    console.error('Update todo error:', error)
    return { error: 'サーバーとの通信に失敗しました' }
  }
}

/**
 * TODO を削除する Server Action
 */
export async function deleteTodoAction(id: number): Promise<DeleteActionResult> {
  try {
    const headers = await getAuthHeaders()

    const response = await fetch(`${API_BASE_URL}/api/todos/${id}`, {
      method: 'DELETE',
      headers,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      return { error: errorData.error || 'TODO の削除に失敗しました' }
    }

    return { success: true }
  } catch (error) {
    console.error('Delete todo error:', error)
    return { error: 'サーバーとの通信に失敗しました' }
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
