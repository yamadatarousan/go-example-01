// Skeleton コンポーネント
// ローディング中のプレースホルダーとして表示される

import { cn } from '@/lib/utils'

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Skeleton({ className, ...props }: SkeletonProps) {
  return (
    <div
      className={cn('animate-pulse rounded-md bg-gray-200', className)}
      {...props}
    />
  )
}

// カード形式のスケルトン（TODO アイテム用）
export function TodoItemSkeleton() {
  return (
    <div className="flex items-center space-x-4 p-4 border rounded-lg">
      <Skeleton className="h-5 w-5 rounded" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-3 w-1/2" />
      </div>
    </div>
  )
}

// TODO リスト全体のスケルトン
export function TodoListSkeleton() {
  return (
    <div className="space-y-3">
      <TodoItemSkeleton />
      <TodoItemSkeleton />
      <TodoItemSkeleton />
      <TodoItemSkeleton />
      <TodoItemSkeleton />
    </div>
  )
}

// 統計カードのスケルトン（ダッシュボード用）
export function StatCardSkeleton() {
  return (
    <div className="p-6 border rounded-lg">
      <Skeleton className="h-4 w-24 mb-2" />
      <Skeleton className="h-8 w-16" />
    </div>
  )
}

// ダッシュボード全体のスケルトン
export function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCardSkeleton />
        <StatCardSkeleton />
        <StatCardSkeleton />
        <StatCardSkeleton />
      </div>
      <div className="space-y-4">
        <Skeleton className="h-6 w-32" />
        <TodoListSkeleton />
      </div>
    </div>
  )
}
