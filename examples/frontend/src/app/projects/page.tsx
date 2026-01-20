// プロジェクト一覧ページ
// Server Component として実装

import Link from 'next/link'
import { getProjectsAction } from './actions'
import { ProjectList } from './project-list'
import { CreateProjectButton } from './create-project-button'
import { Button } from '@/components/ui/button'

export default async function ProjectsPage() {
  const result = await getProjectsAction()
  const projects = result.data || []

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">プロジェクト</h1>
          <nav className="flex items-center space-x-4">
            <Link href="/dashboard" className="text-gray-600 hover:text-gray-900">
              ダッシュボード
            </Link>
            <Link href="/todos" className="text-gray-600 hover:text-gray-900">
              TODO一覧
            </Link>
            <Link href="/settings" className="text-gray-600 hover:text-gray-900">
              設定
            </Link>
          </nav>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* エラー表示 */}
        {result.error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            プロジェクトの取得に失敗しました: {result.error}
          </div>
        )}

        {/* アクションバー */}
        <div className="flex justify-between items-center mb-6">
          <p className="text-gray-600">
            {projects.length} 件のプロジェクト
          </p>
          <CreateProjectButton />
        </div>

        {/* プロジェクト一覧 */}
        {projects.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <svg
              className="mx-auto h-16 w-16 text-gray-400 mb-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              プロジェクトがありません
            </h2>
            <p className="text-gray-600 mb-6">
              新しいプロジェクトを作成して、タスクを整理しましょう
            </p>
            <CreateProjectButton />
          </div>
        ) : (
          <ProjectList projects={projects} />
        )}
      </main>
    </div>
  )
}
