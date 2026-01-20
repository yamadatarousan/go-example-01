// ダッシュボード用 Server Actions
// 統計情報の取得を行う

'use server'

import { fetchWithAuth, mockData, isMockMode } from '@/lib/server-api'

// 統計情報の型定義
export interface TodoStatistics {
  total: number
  by_status: {
    todo: number
    in_progress: number
    done: number
  }
  by_priority: {
    high: number
    medium: number
    low: number
  }
  overdue: number
  due_today: number
  due_this_week: number
}

// 今日のTODOの型
export interface TodayTodo {
  id: number
  name: string
  priority: string
  due_date: string
  status: string
}

/**
 * TODO統計情報を取得する Server Action
 */
export async function getStatisticsAction(): Promise<{
  data?: TodoStatistics
  error?: string
}> {
  // モックモードの場合は仮データを返す
  if (isMockMode()) {
    return { data: mockData.statistics }
  }

  const result = await fetchWithAuth<TodoStatistics>('/api/v1/todos/statistics')
  return result
}

/**
 * 今日のTODO一覧を取得する Server Action
 */
export async function getTodayTodosAction(): Promise<{
  data?: TodayTodo[]
  error?: string
}> {
  const result = await fetchWithAuth<{ todos: TodayTodo[] }>('/api/v1/todos/today')

  if (result.error) {
    return { error: result.error }
  }

  return { data: result.data?.todos || [] }
}

/**
 * 期限切れTODO一覧を取得する Server Action
 */
export async function getOverdueTodosAction(): Promise<{
  data?: TodayTodo[]
  error?: string
}> {
  const result = await fetchWithAuth<{ todos: TodayTodo[] }>('/api/v1/todos/overdue')

  if (result.error) {
    return { error: result.error }
  }

  return { data: result.data?.todos || [] }
}
