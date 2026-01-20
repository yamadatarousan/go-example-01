// 統計カードコンポーネント
// ダッシュボードで統計情報を表示するカード

import { cn } from '@/lib/utils'

type IconType = 'total' | 'done' | 'progress' | 'overdue' | 'high' | 'medium' | 'low'
type Variant = 'default' | 'success' | 'warning' | 'danger'
type Size = 'default' | 'small'

interface StatCardProps {
  title: string
  value: number
  subtitle?: string
  icon: IconType
  variant?: Variant
  size?: Size
}

// アイコンの定義
const icons: Record<IconType, React.ReactNode> = {
  total: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
    </svg>
  ),
  done: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  progress: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  overdue: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
  ),
  high: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
    </svg>
  ),
  medium: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
    </svg>
  ),
  low: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
    </svg>
  ),
}

// バリアントのスタイル
const variantStyles: Record<Variant, { bg: string; icon: string; text: string }> = {
  default: {
    bg: 'bg-gray-50',
    icon: 'text-gray-500',
    text: 'text-gray-900',
  },
  success: {
    bg: 'bg-green-50',
    icon: 'text-green-500',
    text: 'text-green-900',
  },
  warning: {
    bg: 'bg-yellow-50',
    icon: 'text-yellow-500',
    text: 'text-yellow-900',
  },
  danger: {
    bg: 'bg-red-50',
    icon: 'text-red-500',
    text: 'text-red-900',
  },
}

export function StatCard({
  title,
  value,
  subtitle,
  icon,
  variant = 'default',
  size = 'default',
}: StatCardProps) {
  const styles = variantStyles[variant]

  return (
    <div
      className={cn(
        'rounded-lg shadow bg-white overflow-hidden',
        size === 'small' ? 'p-4' : 'p-6'
      )}
    >
      <div className="flex items-center">
        <div className={cn('rounded-lg p-3', styles.bg)}>
          <div className={styles.icon}>{icons[icon]}</div>
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <div className="flex items-baseline">
            <p className={cn('text-2xl font-semibold', styles.text)}>
              {value.toLocaleString()}
            </p>
            {subtitle && (
              <span className="ml-2 text-sm text-gray-500">{subtitle}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
