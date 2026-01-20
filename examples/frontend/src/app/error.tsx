'use client'

// グローバルエラーページ
// 予期しないエラーが発生した場合に表示される
// error.tsx は Client Component である必要がある（エラーバウンダリのため）

import { useEffect } from 'react'
import { Button } from '@/components/ui/button'

interface ErrorProps {
  error: Error & { digest?: string }
  reset: () => void
}

export default function Error({ error, reset }: ErrorProps) {
  useEffect(() => {
    // エラーをログに出力（本番環境では外部サービスに送信することも可能）
    console.error('アプリケーションエラー:', error)
  }, [error])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center p-8 max-w-md">
        <div className="mb-6">
          <svg
            className="mx-auto h-16 w-16 text-red-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          エラーが発生しました
        </h1>
        <p className="text-gray-600 mb-6">
          予期しないエラーが発生しました。問題が続く場合は、サポートにお問い合わせください。
        </p>
        {error.digest && (
          <p className="text-sm text-gray-400 mb-4">
            エラーID: {error.digest}
          </p>
        )}
        <div className="space-x-4">
          <Button onClick={reset} variant="default">
            再試行
          </Button>
          <Button
            onClick={() => (window.location.href = '/')}
            variant="outline"
          >
            ホームに戻る
          </Button>
        </div>
      </div>
    </div>
  )
}
