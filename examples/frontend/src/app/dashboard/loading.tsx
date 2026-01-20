// ダッシュボードローディングページ
// データ取得中に表示される

import { DashboardSkeleton } from '@/components/ui/skeleton'

export default function DashboardLoading() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="h-8 w-40 bg-gray-200 rounded animate-pulse" />
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <DashboardSkeleton />
      </main>
    </div>
  )
}
