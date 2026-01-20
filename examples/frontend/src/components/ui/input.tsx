// shadcn/ui Input コンポーネント
// フォーム入力用の再利用可能なテキストインプット

import * as React from 'react'

import { cn } from '@/lib/utils'

// Input の Props 型定義
// HTML input 要素の全属性を継承
export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

// Input コンポーネント
const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          // ベーススタイル
          'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm',
          // フォーカス時のスタイル
          'ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          // プレースホルダーのスタイル
          'placeholder:text-muted-foreground',
          // ファイル入力時の特別スタイル
          'file:border-0 file:bg-transparent file:text-sm file:font-medium',
          // 無効時のスタイル
          'disabled:cursor-not-allowed disabled:opacity-50',
          // カスタムクラス名を追加
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = 'Input'

export { Input }
