// TODO 一覧ページ
// Server Component: 初期データをサーバーサイドで取得

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { TodoList } from './todo-list'
import { Header } from './header'
import type { Todo } from '@/types'

// API のベース URL
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

// TODO 一覧を取得する関数
async function getTodos(token: string): Promise<Todo[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/todos`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      // キャッシュを無効化（常に最新のデータを取得）
      cache: 'no-store',
    })

    if (!response.ok) {
      if (response.status === 401) {
        // 認証エラーの場合はログインページへ
        return []
      }
      throw new Error('Failed to fetch todos')
    }

    const data = await response.json()
    return data.todos || []
  } catch (error) {
    console.error('Error fetching todos:', error)
    return []
  }
}

export default async function TodosPage() {
  // サーバーサイドで Cookie から JWT トークンを取得
  const cookieStore = await cookies()
  const token = cookieStore.get('token')?.value

  // トークンがない場合はログインページへリダイレクト
  if (!token) {
    redirect('/login')
  }

  // TODO 一覧を取得
  const initialTodos = await getTodos(token)

  return (
    <div className="min-h-screen bg-background">
      {/* ヘッダー（ログアウトボタン付き） */}
      <Header />

      {/* メインコンテンツ */}
      <main className="container mx-auto px-4 py-8 max-w-2xl">
        <h1 className="text-3xl font-bold mb-8">TODO リスト</h1>

        {/* クライアントコンポーネント: TODO の操作を処理 */}
        <TodoList initialTodos={initialTodos} />
      </main>
    </div>
  )
}
