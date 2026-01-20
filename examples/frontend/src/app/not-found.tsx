// 404 Not Found ページ
// 存在しないルートにアクセスした場合に表示される

import Link from 'next/link'
import { Button } from '@/components/ui/button'

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center p-8 max-w-md">
        <div className="mb-6">
          <span className="text-8xl font-bold text-gray-300">404</span>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          ページが見つかりません
        </h1>
        <p className="text-gray-600 mb-6">
          お探しのページは存在しないか、移動した可能性があります。
        </p>
        <div className="space-x-4">
          <Button asChild>
            <Link href="/">ホームに戻る</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/todos">TODO一覧へ</Link>
          </Button>
        </div>
      </div>
    </div>
  )
}
