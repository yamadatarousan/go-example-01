// ダッシュボードページ
// TODO の統計情報を表示する Server Component

import Link from 'next/link'
import { getStatisticsAction, getTodayTodosAction, getOverdueTodosAction } from './actions'
import { StatCard } from './stat-card'
import { TodoPreviewList } from './todo-preview-list'
import { Button } from '@/components/ui/button'

export default async function DashboardPage() {
  // 並列でデータを取得
  const [statsResult, todayResult, overdueResult] = await Promise.all([
    getStatisticsAction(),
    getTodayTodosAction(),
    getOverdueTodosAction(),
  ])

  const stats = statsResult.data
  const todayTodos = todayResult.data || []
  const overdueTodos = overdueResult.data || []

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">ダッシュボード</h1>
          <nav className="flex items-center space-x-4">
            <Link href="/todos" className="text-gray-600 hover:text-gray-900">
              TODO一覧
            </Link>
            <Link href="/projects" className="text-gray-600 hover:text-gray-900">
              プロジェクト
            </Link>
            <Link href="/settings" className="text-gray-600 hover:text-gray-900">
              設定
            </Link>
          </nav>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* エラー表示 */}
        {statsResult.error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            統計情報の取得に失敗しました: {statsResult.error}
          </div>
        )}

        {/* 統計カード */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard
              title="総タスク数"
              value={stats.total}
              icon="total"
            />
            <StatCard
              title="完了済み"
              value={stats.by_status.done}
              subtitle={`${stats.total > 0 ? Math.round((stats.by_status.done / stats.total) * 100) : 0}%`}
              icon="done"
              variant="success"
            />
            <StatCard
              title="進行中"
              value={stats.by_status.in_progress}
              icon="progress"
              variant="warning"
            />
            <StatCard
              title="期限切れ"
              value={stats.overdue}
              icon="overdue"
              variant="danger"
            />
          </div>
        )}

        {/* 優先度別カード */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
            <StatCard
              title="高優先度"
              value={stats.by_priority.high}
              icon="high"
              variant="danger"
              size="small"
            />
            <StatCard
              title="中優先度"
              value={stats.by_priority.medium}
              icon="medium"
              variant="warning"
              size="small"
            />
            <StatCard
              title="低優先度"
              value={stats.by_priority.low}
              icon="low"
              variant="default"
              size="small"
            />
          </div>
        )}

        {/* TODO プレビュー */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 期限切れ TODO */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold text-gray-900">
                期限切れ
                {overdueTodos.length > 0 && (
                  <span className="ml-2 text-sm text-red-500">
                    ({overdueTodos.length}件)
                  </span>
                )}
              </h2>
              <Button asChild variant="ghost" size="sm">
                <Link href="/todos?filter=overdue">すべて表示</Link>
              </Button>
            </div>
            <TodoPreviewList todos={overdueTodos} emptyMessage="期限切れのタスクはありません" />
          </div>

          {/* 今日の TODO */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold text-gray-900">
                今日のタスク
                {todayTodos.length > 0 && (
                  <span className="ml-2 text-sm text-blue-500">
                    ({todayTodos.length}件)
                  </span>
                )}
              </h2>
              <Button asChild variant="ghost" size="sm">
                <Link href="/todos?filter=today">すべて表示</Link>
              </Button>
            </div>
            <TodoPreviewList todos={todayTodos} emptyMessage="今日のタスクはありません" />
          </div>
        </div>

        {/* クイックアクション */}
        <div className="mt-8 flex justify-center space-x-4">
          <Button asChild>
            <Link href="/todos">TODO一覧を見る</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/projects">プロジェクトを管理</Link>
          </Button>
        </div>
      </main>
    </div>
  )
}
