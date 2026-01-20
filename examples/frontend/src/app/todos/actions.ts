// TODO 関連の Server Actions
// サーバーサイドで実行され、API との通信を行う

'use server'

import { cookies } from 'next/headers'
import { fetchWithAuth } from '@/lib/server-api'
import type { Todo } from '@/types'

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
 * TODO を作成する Server Action
 */
export async function createTodoAction(title: string): Promise<TodoActionResult> {
  const result = await fetchWithAuth<Todo>('/api/v1/todos', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })

  if (result.error) {
    return { error: result.error }
  }

  return { todo: result.data }
}

/**
 * TODO を更新する Server Action
 */
export async function updateTodoAction(
  id: number,
  updates: { title?: string; completed?: boolean }
): Promise<TodoActionResult> {
  const result = await fetchWithAuth<Todo>(`/api/v1/todos/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  })

  if (result.error) {
    return { error: result.error }
  }

  return { todo: result.data }
}

/**
 * TODO を削除する Server Action
 */
export async function deleteTodoAction(id: number): Promise<DeleteActionResult> {
  const result = await fetchWithAuth<void>(`/api/v1/todos/${id}`, {
    method: 'DELETE',
  })

  if (result.error) {
    return { error: result.error }
  }

  return { success: true }
}

/**
 * ログアウト Server Action
 * Cookie から JWT トークンを削除する
 */
export async function logoutAction(): Promise<void> {
  const cookieStore = await cookies()
  cookieStore.delete('token')
}
