// グローバルローディングページ
// ページ遷移時に Suspense の fallback として表示される

import { Skeleton } from '@/components/ui/skeleton'

export default function Loading() {
  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        {/* ヘッダー部分のスケルトン */}
        <div className="flex justify-between items-center mb-8">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-10 w-24" />
        </div>

        {/* コンテンツ部分のスケルトン */}
        <div className="space-y-4">
          <Skeleton className="h-12 w-full" />
          <div className="space-y-3">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        </div>
      </div>
    </div>
  )
}
