// プロジェクト詳細ページ
// Server Component として実装

import Link from 'next/link'
import { notFound } from 'next/navigation'
import { getProjectAction, getProjectMembersAction } from '../actions'
import { ProjectHeader } from './project-header'
import { MemberList } from './member-list'
import { Button } from '@/components/ui/button'

interface ProjectDetailPageProps {
  params: Promise<{ id: string }>
}

export default async function ProjectDetailPage({ params }: ProjectDetailPageProps) {
  const { id } = await params
  const projectId = parseInt(id, 10)

  if (isNaN(projectId)) {
    notFound()
  }

  // 並列でデータを取得
  const [projectResult, membersResult] = await Promise.all([
    getProjectAction(projectId),
    getProjectMembersAction(projectId),
  ])

  if (projectResult.error || !projectResult.data) {
    notFound()
  }

  const project = projectResult.data
  const members = membersResult.data || []

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center space-x-4">
            <Link
              href="/projects"
              className="text-gray-500 hover:text-gray-700"
            >
              <svg
                className="w-6 h-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
            </Link>
            <h1 className="text-2xl font-bold text-gray-900">{project.name}</h1>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* メイン情報 */}
          <div className="lg:col-span-2 space-y-6">
            {/* プロジェクト情報カード */}
            <div className="bg-white rounded-lg shadow p-6">
              <ProjectHeader project={project} />

              {project.description && (
                <div className="mt-4">
                  <h3 className="text-sm font-medium text-gray-500 mb-2">説明</h3>
                  <p className="text-gray-700 whitespace-pre-wrap">
                    {project.description}
                  </p>
                </div>
              )}

              <div className="mt-6 pt-4 border-t border-gray-200">
                <div className="flex items-center space-x-4 text-sm text-gray-500">
                  <span>
                    作成日:{' '}
                    {new Date(project.created_at).toLocaleDateString('ja-JP')}
                  </span>
                  <span>
                    更新日:{' '}
                    {new Date(project.updated_at).toLocaleDateString('ja-JP')}
                  </span>
                </div>
              </div>
            </div>

            {/* TODO一覧へのリンク */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                タスク
              </h2>
              <p className="text-gray-600 mb-4">
                このプロジェクトに関連するタスクを管理します
              </p>
              <Button asChild>
                <Link href={`/todos?project=${projectId}`}>
                  タスク一覧を表示
                </Link>
              </Button>
            </div>
          </div>

          {/* サイドバー */}
          <div className="space-y-6">
            {/* メンバー一覧 */}
            <div className="bg-white rounded-lg shadow p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-semibold text-gray-900">
                  メンバー
                </h2>
                <span className="text-sm text-gray-500">
                  {members.length} 人
                </span>
              </div>
              <MemberList members={members} projectId={projectId} />
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
